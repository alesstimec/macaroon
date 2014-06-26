package checkers

func ExpiresBefore(t time.Time) bakery.Caveat {
	return ThirdParty("expires", t)
}

type StructuredCaveat struct {
	Identifier string
	Args       []interface{}
}

func FirstParty(identifier string, args ...interface{}) bakery.Caveat {
	return ThirdParty("", identifier, args...)
}

func ThirdParty(location, identifier string, args ...interface{}) bakery.Caveat {
	c := StructuredCaveat{
		Identifier: identifier,
		Args:       args,
	}
	data, err := json.Marshal(c)
	if err != nil {
		panic(fmt.Errorf("cannot marshal %#v: %v", c, err))
	}
	return bakery.Caveat{
		Location:  location,
		Condition: string(data),
	}
}

func ParseCaveat(cav string) (*StructuredCaveat, error) {
	if cav == "" {
		// TODO(rog) Or should we just return an empty caveat?
		return nil, fmt.Errorf("empty caveat")
	}
	if cav[0] == '{' {
		var c StructuredCaveat
		if err := json.Unmarshal([]byte(cav), &c); err != nil {
			return fmt.Errorf("cannot unmarshal structured caveat: %v", err)
		}
		return &c
	}
	toks := strings.SplitN(cav, " ", 3)
	return StructuredCaveat{
		Identifier: toks[0],
		Args:       toks[1:],
	}, nil
}
