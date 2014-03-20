package retr

import (
	"errors"
	. "github.com/trapped/gomaild/parsers/textual"
	"github.com/trapped/gomaild/processors/pop3/message"
	. "github.com/trapped/gomaild/processors/pop3/session"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

func Process(session *Session, c Statement) (string, error) {
	errorslice := []string{}
	result := ""
	goto checks

returnerror:
	if len(errorslice) != 0 {
		result = strings.Join(errorslice, ", ")
		return "", errors.New(result)
	}

checks:
	if session.State != TRANSACTION {
		errorslice = append(errorslice, "wrong session state")
	}
	if !session.Authenticated {
		errorslice = append(errorslice, "not authenticated")
	}
	if session.Username == "" {
		errorslice = append(errorslice, "user can't be empty")
	}
	if len(c.Arguments) > 2 {
		errorslice = append(errorslice, "too many arguments")
	}
	if len(c.Arguments) == 1 {
		errorslice = append(errorslice, "message ID can't be empty")
	}

	if len(errorslice) != 0 {
		goto returnerror
	}

	log.Println("POP3:", "RETR command issued by", session.RemoteEP, "with", session.Username)

	messages := message.Index(session)

	for _, v := range messages {
		if strconv.Itoa(v.ID) == c.Arguments[1] {
			result = strconv.Itoa(int(v.File.Size())) + " octets\r\n"
			file, err := ioutil.ReadFile(v.Path)
			if err != nil {
				errorslice = append(errorslice, "error reading the message")
				goto returnerror
			}
			result += string(file)
			result += "\r\n."

			session.Retrieved = append(session.Retrieved, v)
			break
		}
	}
	if result == "" {
		errorslice = append(errorslice, "no such message; "+strconv.Itoa(len(messages))+" messages in maildrop")
		goto returnerror
	}

	return result, nil
}
