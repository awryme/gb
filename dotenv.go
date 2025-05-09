package main

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/awryme/slogf"
)

func DotenvLoad(log slog.Handler, name string) error {
	const defaultFile = ".env"
	const gbFileFmt = ".%s.gb"
	const envFileFmt = ".%s.env"
	logf := slogf.New(log)
	err := dotenvReadFile(defaultFile, false)
	switch {
	case err == nil:
		logf("using .env")
	case errors.Is(err, os.ErrNotExist):
		// pass
	default:
		return fmt.Errorf("load .env: %w", err)
	}

	if name == "" {
		return nil
	}

	file := name
	err1 := dotenvReadFile(file, true)
	if err1 == nil {
		logf("using dotenv file", slog.String("name", file))
		return nil
	}

	file = fmt.Sprintf(gbFileFmt, name)
	err2 := dotenvReadFile(file, true)
	if err2 == nil {
		logf("using dotenv file", slog.String("name", file))
		return nil
	}

	file = fmt.Sprintf(envFileFmt, name)
	err3 := dotenvReadFile(file, true)
	if err3 == nil {
		logf("using dotenv file", slog.String("name", file))
		return nil
	}

	err = errors.Join(err1, err2, err3)
	return fmt.Errorf("load dotenv file: %w", err)
}

func dotenvReadFile(filename string, override bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var idx int64
	for scanner.Scan() {
		idx++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		strs := strings.SplitN(line, "=", 2)
		if len(strs) != 2 {
			return fmt.Errorf("incorrect env line #%d: '%s' cannot be parsed as env format KEY=VALUE", idx, line)
		}
		key, value := strs[0], strs[1]
		if !override {
			if _, ok := os.LookupEnv(key); ok {
				continue
			}
		}
		err := os.Setenv(key, value)
		if err != nil {
			return err
		}
	}
	return scanner.Err()
}
