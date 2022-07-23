package lib

// github.com/pilinux/gorest
// The MIT License (MIT)
// Copyright (c) 2022 pilinux

import "github.com/alexedwards/argon2id"

// HashPassConfig - params for argon2id
type HashPassConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// HashPass - securely hash passwords using Argon2id
func HashPass(config HashPassConfig, pass string) (string, error) {
	params := &argon2id.Params{
		Memory:      config.Memory * 1024, // the amount of memory used by the Argon2 algorithm (in kibibytes)
		Iterations:  config.Iterations,    // the number of iterations (or passes) over the memory
		Parallelism: config.Parallelism,   // the number of threads (or lanes) used by the algorithm
		SaltLength:  config.SaltLength,    // length of the random salt. 16 bytes is recommended for password hashing
		KeyLength:   config.KeyLength,     // length of the generated key (or password hash). 16 bytes or more is recommended
	}
	h, err := argon2id.CreateHash(pass, params)
	if err != nil {
		return "", err
	}
	return h, err
}
