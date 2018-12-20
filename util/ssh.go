package util

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/ssh"
)

//
// sshConnectToHost connects to a given host with the given password.
//

func sshConnectToHost(host, port, user, password string) (*ssh.Client, *ssh.Session, error) {

	keyboardInteractiveChallenge := func(
		user,
		instruction string,
		questions []string,
		echos []bool,
	) (answers []string, err error) {
		if len(questions) == 0 {
			return []string{}, nil
		}
		return []string{password}, nil
	}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.KeyboardInteractive(keyboardInteractiveChallenge),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConfig.SetDefaults()

	fullHost := fmt.Sprintf("%s:%s", host, port)
	client, err := ssh.Dial("tcp", fullHost, sshConfig)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		return nil, nil, err
	}

	return client, session, nil
}
//Excutescript Connect to remote host and excute script on it,return result which is splited by '\n'
func Excutescript(host, port, user, password, cmd string) ([]string, error) {
	client, session, err := sshConnectToHost(host, port, user, password)
	if err != nil {
		return nil, err
	}
	services, err := session.Output(cmd)
	if err != nil {
		return nil, err
	}
	strser := string(services[:])
	servicelist := strings.Split(strser, "\n")
	defer client.Close()
	return servicelist, err
}

// Dealstr deal with a array of string through regexp
func Dealstr(servicelist []string) ([]string, []string) {
	servname := make([]string, len(servicelist)-1)
	status := make([]string, len(servicelist)-1)
	for ser, sl := range servicelist {
		service := string(sl)
		reg := regexp.MustCompile(`^\w+`)
		str := reg.FindAllString(service, -1)
		if len(str) > 0 {
			servname[ser] = str[0]
		}
		reg2 := regexp.MustCompile(`\s\b(.+)$`)
		str2 := reg2.FindAllString(service, -1)
		var tempstring string
		if len(str2) > 0 {
			for i := 0; i < len(str2); i++ {
				tempstring += str2[i]
			}
			reg3 := regexp.MustCompile(`\b(.+)$`)
			str3 := reg3.FindAllString(tempstring, -1)
			for i := 0; i < len(str3); i++ {
				status[ser] += str3[i]
			}
		}

	}
	return servname, status
}
