package server

import (
	"fmt"
	"net"
	"os"

	"github.com/chzyer/flagly"
	"github.com/chzyer/logex"
	"github.com/chzyer/readline"
	"github.com/google/shlex"
)

var Slogan = `
   _  _______  ________
  / |/ / __/ |/_/_  __/
 /    / _/_>  <  / /   
/_/|_/___/_/|_| /_/    
`

type Shell struct {
	sock string
	conn net.Listener
	svr  *Server
}

func NewShell(svr *Server, sock string) (*Shell, error) {
	ln, err := net.Listen("unix", sock)
	if err != nil {
		return nil, err
	}
	sh := &Shell{
		sock: sock,
		conn: ln,
		svr:  svr,
	}
	return sh, nil
}

func (s *Shell) handleConn(conn net.Conn) {
	defer conn.Close()

	sh := &ShellCLI{}
	fset, err := flagly.Compile("", sh)
	if err != nil {
		logex.Info(err)
		return
	}

	cfg := readline.Config{
		Prompt:       "server> ",
		AutoComplete: readline.SegmentAutoComplete(fset.Completer()),
	}
	rl, err := readline.HandleConn(cfg, conn)
	if err != nil {
		return
	}
	defer rl.Close()

	fset.Context(rl, s.svr)
	fmt.Fprintln(rl, Slogan)
	for {
		command, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(command) == 0 {
				break
			} else {
				continue
			}
		} else if err != nil {
			break
		}

		args, err := shlex.Split(command)
		if err != nil {
			continue
		}

		if err := fset.Run(args); err != nil {
			fmt.Fprintln(rl.Stderr(), err)
			continue
		}
	}
}

func (s *Shell) loop() {
	for {
		conn, err := s.conn.Accept()
		if err != nil {
			break
		}
		go s.handleConn(conn)
	}
}

func (s *Shell) Close() {
	s.conn.Close()
	os.Remove(s.sock)
}

type ShellCLI struct {
	Help  flagly.CmdHelp `flagly:"handler"`
	User  ShellUser      `flagly:"handler"`
	Debug *ShellDebug    `flagly:"handler"`
	Dchan *Dchan         `flagly:"handler"`
}
