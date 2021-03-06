//Implements the PASS command.
package pass

import (
	"errors"
	"github.com/trapped/gomaild/config"
	"github.com/trapped/gomaild/locker"
	"github.com/trapped/gomaild/mailboxes"
	. "github.com/trapped/gomaild/parsers/textual"
	. "github.com/trapped/gomaild/processors/pop3/session"
	"log"
	"strconv"
	"strings"
)

//Processes the PASS command.
func Process(session *Session, c Statement) (string, error) {
	errorslice := []string{}
	result := ""
	goto checks

returnerror:
	session.Username = ""
	session.Password = ""
	if len(errorslice) != 0 {
		result = strings.Join(errorslice, ", ")
		return "", errors.New(result)
	}

checks:
	if !config.Configuration.POP3.EnableUSER {
		errorslice = append(errorslice, "command not available")
		goto returnerror
	}
	if session.State != AUTHORIZATION {
		errorslice = append(errorslice, "wrong session state")
	}
	if session.Authenticated {
		errorslice = append(errorslice, "already authenticated")
	}
	if session.Username == "" {
		errorslice = append(errorslice, "use command USER first")
	}
	if session.Password != "" {
		errorslice = append(errorslice, "session password already set")
	}
	if len(c.Arguments) == 1 {
		errorslice = append(errorslice, "password can't be empty")
	}
	if len(c.Arguments) > 2 {
		errorslice = append(errorslice, "too many arguments")
	}

	if len(errorslice) != 0 {
		goto returnerror
	}

	log.Println("POP3:", "PASS command issued by", session.RemoteEP, "with", session.Username, "and `"+session.Password+"`")

	password, exists := mailboxes.GetUser(session.Username)
	if exists != nil {
		errorslice = append(errorslice, config.Configuration.POP3.PasswordInvalidMessage)
		goto returnerror
	}

	if password != c.Arguments[1] {
		errorslice = append(errorslice, config.Configuration.POP3.PasswordInvalidMessage)
		goto returnerror
	}

	lockerr := locker.Lock(mailboxes.GetMailbox(session.Username))
	if lockerr != nil {
		errorslice = append(errorslice, "[IN-USE] maildrop "+lockerr.Error())
		goto returnerror
	}

	session.Password = c.Arguments[1]
	session.Authenticated = true
	session.State = TRANSACTION
	count, octets := mailboxes.Stat(session.Username, false)
	result = session.Username + "'s maildrop has " + strconv.Itoa(count) + " messages (" + strconv.Itoa(octets) + " octets)"

	return result, nil
}
