package main

type tokenRegistry struct {
	symTable map[string]*token
}

func (self *tokenRegistry) token(sym string, value string, line int, col int) *token {
	return &token{
		sym:          sym,
		value:        value,
		line:         line,
		col:          col,
		bindingPower: self.symTable[sym].bindingPower,
		nud:          self.symTable[sym].nud,
		led:          self.symTable[sym].led,
		std:          self.symTable[sym].std,
	}
}

func (self *tokenRegistry) defined(sym string) bool {
	if _, ok := self.symTable[sym]; ok {
		return true
	}
	return false
}

func (self *tokenRegistry) register(sym string, bp int, nud nudFn, led ledFn, std stdFn) {
	if val, ok := self.symTable[sym]; ok {
		if nud != nil && val.nud == nil {
			val.nud = nud
		}
		if led != nil && val.led == nil {
			val.led = led
		}
		if std != nil && val.std == nil {
			val.std = std
		}
		if bp > val.bindingPower {
			val.bindingPower = bp
		}
	} else {
		self.symTable[sym] = &token{bindingPower: bp, nud: nud, led: led, std: std}
	}
}

// an infix token has two children, the exp on the left and the one that follows
func (self *tokenRegistry) infix(sym string, bp int) {
	self.register(sym, bp, nil, func(t *token, p *parser, left *token) *token {
		t.children = append(t.children, left)
		t.children = append(t.children, p.expression(t.bindingPower))
		return t
	}, nil)
}

func (self *tokenRegistry) infixLed(sym string, bp int, led ledFn) {
	self.register(sym, bp, nil, led, nil)
}

func (self *tokenRegistry) infixRight(sym string, bp int) {
	self.register(sym, bp, nil, func(t *token, p *parser, left *token) *token {
		t.children = append(t.children, left)
		t.children = append(t.children, p.expression(t.bindingPower-1))
		return t
	}, nil)
}

func (self *tokenRegistry) infixRightLed(sym string, bp int, led ledFn) {
	self.register(sym, bp, nil, led, nil)
}

// a prefix token has a single children, the expression that follows
func (self *tokenRegistry) prefix(sym string) {
	self.register(sym, 0, func(t *token, p *parser) *token {
		t.children = append(t.children, p.expression(100))
		return t
	}, nil, nil)
}

func (self *tokenRegistry) prefixNud(sym string, nud nudFn) {
	self.register(sym, 0, nud, nil, nil)
}

func (self *tokenRegistry) stmt(sym string, std stdFn) {
	self.register(sym, 0, nil, nil, std)
}

func (self *tokenRegistry) symbol(sym string) {
	self.register(sym, 0, func(t *token, p *parser) *token { return t }, nil, nil)
}

func (self *tokenRegistry) consumable(sym string) {
	self.register(sym, 0, nil, nil, nil)
}
