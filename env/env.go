package env

import (
	"bufio"
	"os"
	"path"
	"strings"
)

type Environment map[string]string

const COMMENT = '#'

func Load(envFile string) (func(), error) {
	executable, err := os.Executable()
	if err != nil {
		return func() {}, err
	}

	envPath := path.Join(path.Dir(executable), envFile)
	file, err := os.Open(envPath)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	oldEnv := make(Environment)
	for scanner.Scan() {
		line := scanner.Text()

		index := strings.Index(line, "=")
		if index == -1 || line[0] == COMMENT {
			continue
		}
		key := strings.TrimSpace(line[:index])
		value := strings.TrimSpace(line[index+1:])

		oldEnv[key] = os.Getenv(key)
		os.Setenv(key, value)
	}

	file.Close()

	return func() {
		restore(&oldEnv)
	}, nil
}

func restore(env *Environment) {
	for key, value := range *env {
		if len(value) == 0 {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}
