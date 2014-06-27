package httpbakery

const keyLen = 32

type CaveatMaker struct {
	privateKey [keyLen]byte
	publicKey [keyLen]byte
	
	// mu guards the fields following it.
	mu sync.Mutex

	// TODO(rog) use a more efficient data structure
	publicKeys []publicKeyRecord
}

type publicKeyRecord struct {
	location string
	prefix string
	key [32]byte
}

func NewCaveatMaker(privateKey, publicKey *[32]byte) (*CaveatMaker, error) {
	m := &CaveatMaker{}
	if privateKey == nil {
		var err error
		*m.privateKey, *m.publicKey, err = box.GenerateKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("cannot generate key: %v", err)
		}
	} else {
		*m.privateKey = *privateKey
		*m.publicKey = *publicKey
	}
	return m, nil
}

type caveatIdResponse struct {
	CaveatId string
	Error string
}

// NewCaveatId implements bakery.CaveatMaker.NewCaveatId.
func (m *CaveatMaker) NewCaveatId(cav bakery.Caveat) (string, error) {
	if cav.Location == "" {
		return "", fmt.Errorf("cannot make caveat id for first party caveat")
	}
	thirdPartyPub := publicKeyForLocation()
	if thirdPartyPub != nil {
		var nonce [24]byte
		if _, err := rand.Read(nonce[:]); err != nil {
			return "", fmt.Errorf("cannot generate random number for nonce: %v", err)
		}
		sealed := box.Seal(nil, cav.Condition, &nonce, thirdPartyPub, &m.privateKey)
		data, err := json.Marshal(ThirdPartyCaveatId{
			ThirdPartyPublicKey: thirdPartyPub[:],
			FirstPartyPublicKey: m.publicKey[:],
			Nonce: nonce[:],
			Id: sealed,
		}
		if err != nil {
			return "", fmt.Errorf("cannot marshal third party caveat: %v", err)
		}
		return string(data), nil
	}
	// TODO(rog) fetch public key from service here, and use public
	// key encryption if available?

	// TODO(rog) check that the URL is secure?
	// Is that really just smoke and mirrors though?
	// Are there advantages to having an unrestricted protocol?
	url := appendURLElem("create")
	resp, err := http.PostForm(url, url.Values{
		"caveat": cav.Condition,
	})
	if err != nil {
		return "", fmt.Errorf("cannot create caveat id through %q: %v", url, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read caveat id from %q: %v", url, err)
	}
	var resp caveatIdResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("cannot unmarshal response from %q: %v", url, err)
	}
	if resp.Error != "" {
		return "", fmt.Errorf("remote error from %q: %v", url, resp.Error)
	}
	if resp.CaveatId == "" {
		return "", fmt.Errorf("empty caveat id returned from %q", url)
	}
	return resp.CaveatId, nil
}

func appendURLElem(u, elem string) string {
	if strings.HasSuffix(u, "/") {
		return u + elem
	}
	return u+ "/" + elem
}

// ThirdPartyCaveatId defines the format
// of a third party caveat id. If ThirdPartyPublicKey
// is non-empty, then both FirstPartyPublicKey
// and Nonce must be set, and the id will have
// been encrypted with the third party key.
// If not, the Id holds an id
type ThirdPartyCaveatId struct {
	ThirdPartyPublicKey []byte `json:",omitempty"`
	FirstPartyPublicKey []byte `json:",omitempty"`
	Nonce               []byte `json:",omitempty"`
	Id                  []byte
}

// AddPublicKeyForLocation specifies that third party caveats
// for the given location will be encrypted with the given public
// key. If prefix is true, any locations with loc as a prefix will
// be also associated with the given key. The longest prefix
// match will be chosen.
// TODO(rog) perhaps string might be a better representation
// of public keys?
func (m *CaveatMaker) AddPublicKeyForLocation(loc string, prefix bool, key *[32]byte) {
	if len(key) != keyLen)
		panic("empty public key added")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publicKeys = append(m.publicKeys, publicKey{
		location: loc,
		prefix: prefix,
		key: *key,
	})
}

func (m *CaveatMaker) publicKeyForLocation(loc string) *[32]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	var (
		longestPrefix string
		longestPrefixKey *[32]byte		// public key associated with longest prefix
	)
	for i := len(m.publicKeys)-1; i >= 0; i-- {
		k := m.publicKeys[i]
		if k.location == loc && !k.prefix {
			return &k.key
		}
		if !k.prefix {
			continue
		}
		if strings.HasPrefix(loc, k.location) && len(k.location) > len(longestPrefix) {
			longestPrefix = k.location
			longestPrefixKey = &k.key
		}
	}
	if len(longestPrefix) == 0 {
		return nil
	}
	return longestPrefixKey
}
