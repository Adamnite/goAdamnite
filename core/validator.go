package core

type ValidatorConfig struct {
	ValidatorCount uint32 // The total numbers of validator on top tier
}

var DefaultValidatorConfig = ValidatorConfig{
	ValidatorCount: 27,
}
