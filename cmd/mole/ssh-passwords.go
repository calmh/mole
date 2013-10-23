package main

type password string

func (p password) Password(user string) (string, error) {
	return string(p), nil
}

type challenge string

func (c challenge) Challenge(user, instruction string, questions []string, echos []bool) ([]string, error) {
	answers := make([]string, len(questions))
	for i := range answers {
		answers[i] = string(c)
	}
	return answers, nil
}
