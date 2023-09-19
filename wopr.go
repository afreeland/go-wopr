package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

type Screen int

// These screens represent each state in our terminal app
const (
	LOGON Screen = iota
	GAMES
	GREETING
	WELLBEING
	EXPLANATION
	PLAY_GAME
	PLAY_GAME_VERIFY
	GLOBAL_THERMONOCULEAR_WAR
	STRIKE_COMMAND
	STRIKE_COMMAND_2
)

type ClientState struct {
	Authenticated bool
	Screen
}

// Store a list of tcpClients and if they are active or not
var tcpClients = make(map[net.Conn]ClientState)

// Some common characters and ASCII codes
const (
	FULL_BLOCK         = "█"
	UNDERLINE          = "-"
	CLEAR              = "\u001B[2J"
	CLEAR_AND_RESTORE  = "\033[u\033[K"
	CURSOR_UP          = "\033[A"
	CLEAR_CURRENT_LINE = "\033[K"
	CURRENT_POSITION   = "\033[s"

	US = `
    ,------~~v,                
    |'         Ż\   ,__/Ż||    
   /             \,/     /     
   |                    /      
   \                   |       
    \                 /        
     ^Ż~_            /         
         '~~,  ,Ż~Ż\ \         
             \/     \/         
                               
    `

	SOVIET_UNION = `
              _--^\
            _/    /,_
   ,,   ,,/^      Ż  vŻv-__
   |'~^Ż                   Ż\
 _/                     _  /^
/                   ,~~^/|ŻŻ
|          __,,  v__\   \/
 ^~       /    ~Ż  //
   \~,  ,/         Ż
      ~~
    `
)

var currentRow int = 0

var serverMode bool = false

// This interface lets us support different connection types
// ConsoleMessenger - For terminal app (ie ./wopr)
// NetworkMessenger - TCP connections on 2000 by default (ie ./wopr --server)
// TODO: WebMessenger - WebSockets to UI
type Messenger interface {
	Disconnect()
	Send(msg string)
	ScanSupport() bool
	UpdateState(scr Screen)
}

type ConsoleMessenger struct{}

func (c ConsoleMessenger) Send(msg string) {
	// Since we are dealing with a terminal, we can simply print
	fmt.Print(msg)
}

// Disconnect for our terminal app feels unnecessary, we could kill process but
// feels nicer to just leave it at the LOGON: prompt
func (c ConsoleMessenger) Disconnect() {}

// ScanSupport is set to true when we are in terminal since we can accept user input through scan
func (c ConsoleMessenger) ScanSupport() bool { return true }

func (c ConsoleMessenger) UpdateState(scr Screen) {}

type NetworkMessenger struct {
	client net.Conn
}

// NewNetworkMessenger allows easy instantiation of a NetworkMessenger while also tracking our
// tcp client for updating throughout various stages, as well as disconnecting
func NewNetworkMessenger(conn net.Conn) *NetworkMessenger {
	return &NetworkMessenger{client: conn}
}

// Send will write the the client connection
func (nm NetworkMessenger) Send(msg string) {
	nm.client.Write([]byte(msg))
}

// Disconnect will
func (nm NetworkMessenger) Disconnect() {
	nm.Send("\033[0m")
	delete(tcpClients, nm.client)
	nm.client.Close()
}

func (nm NetworkMessenger) ScanSupport() bool { return false }

func (nm NetworkMessenger) UpdateState(scr Screen) {
	state := tcpClients[nm.client]
	state.Screen = scr
	tcpClients[nm.client] = state
}

func main() {
	// Define a flag for the server option
	serverPtr := flag.Bool("server", false, "Enable server mode to actually listen to clients")

	// Parse the command-line arguments
	flag.Parse()

	if *serverPtr {
		// Server mode enabled
		// Handle actual TCP connections and allows us to interact with clients

		serverMode = true

		WOPR_PORT := 2000
		if os.Getenv("WOPR_PORT") != "" {
			port, err := strconv.Atoi(os.Getenv("WOPR_PORT"))
			if err != nil {
				panic(err)
			}
			WOPR_PORT = port
		}

		// Create a listener for TCP traffic on our WOPR_PORT
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", WOPR_PORT))
		if err != nil {
			panic(err)
		}

		for {
			client, err := listener.Accept()
			if err != nil {
				continue
			}
			tcpClients[client] = ClientState{
				Authenticated: false,
				Screen:        LOGON,
			}
			conn := NewNetworkMessenger(client)
			connectClient(conn)
		}
	} else {
		initWOPR(&ConsoleMessenger{})
	}

}

func connectClient(nMsgr *NetworkMessenger) {
	// Kick off our application
	initWOPR(nMsgr)

	// Infinite loop listening to our client for input
	for {
		buf := make([]byte, 4096)
		numBytes, err := nMsgr.client.Read(buf)
		if numBytes == 0 || err != nil {
			return
		}

		// If we get telnet commands, ignore them
		if buf[0] == 255 {
			continue
		}

		// Take the input we recieved and send them to our handler
		handleInput(nMsgr, tcpClients[nMsgr.client].Screen, string(buf[0:numBytes]))
	}
}

func handleInput(msgr Messenger, screen Screen, msg string) {
	// Clean up and normalize our data..especially since it comes from multiple sources
	msg = stripInput(msg)

	// We are only going to fmt print when we are in server mode
	// That way we can conveniently see the input being received =)
	// Otherwise, if we were in console mode it would mess up rendering
	if !msgr.ScanSupport() {
		fmt.Println(msg)
	}
	switch screen {
	// This screen state is the initial state that drives navigation
	case LOGON:
		switch msg {
		case "JOSHUA":
			// They entered correct password and we can move them to another state
			greeting(msgr)
		case "HELP", "HELP LOGON":
			help(msgr)
		case "HELP GAMES":
			// We are navigating to a state that accepts input
			// Update our state so we know how to handle it upon data entry

			helpGames(msgr)
		default:
			notRecognized(msgr)
			return
		}

		// "HELP GAMES" will land you here
	case GAMES:
		switch msg {
		case "LIST GAMES":
			listGames(msgr)
		default:
			// If they dont list games, take them back to LOGON state
			logon(msgr)
		}

		// After successful login, they are directed to a GREETING
		// which we need to listen for any input and move them along the converations
	case GREETING:
		wellbeing(msgr)

		// After asking about wellbeing move them along in the conversation
	case WELLBEING:
		explanation(msgr)

		// After prompting for why account was deleted, they are offered to play a game
	case EXPLANATION:
		playGame(msgr)
		// See if they really want to play THAT game
	case PLAY_GAME:
		playGameVerify(msgr)

		// ALright, looks like Global Thermonuclear War after all...
	case PLAY_GAME_VERIFY:
		globalThermoNuclearWar(msgr)

	case GLOBAL_THERMONOCULEAR_WAR:
		strikeCommands(msgr)

	case STRIKE_COMMAND:
		strikeCommandDos(msgr)

	case STRIKE_COMMAND_2:
		// TODO
	default:
	}

}

func stripInput(input string) string {
	input = strings.Replace(input, "\r", "", -1)
	input = strings.Replace(input, "\n", "", -1)
	input = strings.ToUpper(input)

	return input
}

func initWOPR(msgr Messenger) {
	// Set text color to light blue
	msgr.Send("\x1b[38;5;51m")

	// Server identification
	identifyServer(msgr)
	logon(msgr)

	// Reset text color to default
	msgr.Send("\033[0m")
}

func identifyServer(msgr Messenger) {
	msgr.Send(CLEAR)
	setCursorPosition(0, 0, msgr)
	identificationSpeed := time.Millisecond * 15
	curWidth := getTerminalWidth()

	printBlocks := func() {
		// In the movie the machine draws full block empty spaces and does not 'identify' itself
		// Let's try to replicate that effect...
		for i := 0; i < curWidth; i++ {
			// \r - moves the cursor to the start of the line each time
			msgr.Send(fmt.Sprintf("\r%*s", i, FULL_BLOCK))
			time.Sleep(identificationSpeed)
		}
		msgr.Send(CLEAR_AND_RESTORE)
	}

	printBlocks()
	bumpLine(msgr)
	printBlocks()
}

func logon(msgr Messenger) {
	msgr.UpdateState(LOGON)
	bumpLine(msgr)
	animateMessage(msgr, "LOGON: ")

	if msgr.ScanSupport() {
		login := scan()
		handleInput(msgr, LOGON, login)
	}
}

func help(msgr Messenger) {
	bumpLine(msgr)
	animateMessage(msgr, "HELP NOT AVAILABLE")
	bumpLine(msgr)
	logon(msgr)
}

func helpGames(msgr Messenger) {
	msgr.UpdateState(GAMES)
	bumpLine(msgr)
	animateMessage(msgr, `'GAMES' REFERS TO MODELS, SIMULATIONS AND GAMES`)
	bumpLine(msgr)
	animateMessage(msgr, "WHICH HAVE TACTICAL AND STRATEGIC APPLICATIONS.")
	bumpLine(msgr, 2)

	if msgr.ScanSupport() {
		games := scan()
		handleInput(msgr, GAMES, games)
	}

}

func listGames(msgr Messenger) {
	games := []string{
		"FALKEN'S MAZE",
		"BLACK JACK",
		"GIN RUMMY",
		"HEARTS",
		"BRIDGE",
		"CHECKERS",
		"CHESS",
		"POKER",
		"FIGHTER COMBAT",
		"GUERILLA ENGAGEMENT",
		"DESERT WARFARE",
		"AIR-TO-GROUND ACTIONS",
		"THEATERWIDE TACTICAL WARFARE",
		"THEATERWIDE BIOTOXIC AND CHEMICAL WARFARE",
		"GLOBAL THERMONUCLEAR WAR",
	}
	bumpLine(msgr)

	for i, g := range games {
		// Global Thermonuclear war gets its own space separation
		if i == len(games)-1 {
			bumpLine(msgr)
		}

		animateMessage(msgr, g)
		bumpLine(msgr)

		// Global Thermonuclear war gets its own space separation
		if i == len(games)-1 {
			bumpLine(msgr)
		}
	}

	logon(msgr)

}

func greeting(msgr Messenger) {
	msgr.UpdateState(GREETING)
	msgr.Send(CLEAR)
	setCursorPosition(0, 0, msgr)
	animateMessage(msgr, "GREETINGS PROFESSOR FALKEN.")
	bumpLine(msgr, 2)

	if msgr.ScanSupport() {
		greetings := scan()
		handleInput(msgr, GREETING, greetings)
	}
}

func wellbeing(msgr Messenger) {
	msgr.UpdateState(WELLBEING)
	bumpLine(msgr, 2)
	animateMessage(msgr, "HOW ARE YOU FEELING TODAY?")
	bumpLine(msgr, 2)

	if msgr.ScanSupport() {
		feeling := scan()
		handleInput(msgr, WELLBEING, feeling)
	}
}

func explanation(msgr Messenger) {
	msgr.UpdateState(EXPLANATION)
	bumpLine(msgr, 2)
	animateMessage(msgr, "EXCELLENT. IT'S BEEN A LONG TIME. CAN YOU EXPLAIN")
	bumpLine(msgr)
	animateMessage(msgr, "THE REMOVAL OF YOUR USER ACCOUNT NUMBER ON 6/23/73?")
	bumpLine(msgr, 2)

	if msgr.ScanSupport() {
		explanation := scan()
		handleInput(msgr, EXPLANATION, explanation)
	}
}

func playGame(msgr Messenger) {
	msgr.UpdateState(PLAY_GAME)
	bumpLine(msgr, 2)
	animateMessage(msgr, "YES, THEY DO. SHALL WE PLAY A GAME?")
	bumpLine(msgr, 2)

	if msgr.ScanSupport() {
		playGame := scan()
		handleInput(msgr, PLAY_GAME, playGame)
	}
}

func playGameVerify(msgr Messenger) {
	msgr.UpdateState(PLAY_GAME_VERIFY)
	bumpLine(msgr, 2)
	animateMessage(msgr, "WOULDN'T YOU PREFER A GOOD GAME OF CHESS?")
	bumpLine(msgr, 2)

	if msgr.ScanSupport() {
		verify := scan()
		handleInput(msgr, PLAY_GAME_VERIFY, verify)
	}
}

func animateMessage(msgr Messenger, msg string, dur ...time.Duration) {
	speed := time.Millisecond * 20

	if len(dur) > 0 {
		speed = dur[0]
	}
	for i := 0; i < len(msg); i++ {
		msgr.Send(fmt.Sprintf("%c", msg[i]))
		time.Sleep(speed)
	}
}

func globalThermoNuclearWar(msgr Messenger) {
	msgr.UpdateState(GLOBAL_THERMONOCULEAR_WAR)
	animateMessage(msgr, "Fine.")
	msgr.Send(CLEAR)
	setCursorPosition(0, 0, msgr)

	usLines := strings.Split(US, "\n")
	sovietLines := strings.Split(SOVIET_UNION, "\n")

	maxLines := len(usLines)
	if len(sovietLines) > maxLines {
		maxLines = len(sovietLines)
	}

	// Format and print the ASCII arts side by side
	for i := 0; i < maxLines; i++ {
		// Prepare the lines from both arts for printing
		line1 := ""
		if i < len(usLines) {
			line1 = usLines[i]
		}
		line2 := ""
		if i < len(sovietLines) {
			line2 = sovietLines[i]
		}

		// Use fmt.Printf to print the lines side by side
		animateMessage(msgr, fmt.Sprintf("%-40s  %-40s\n", line1, line2), time.Millisecond*5)
	}
	bumpLine(msgr)

	animateMessage(msgr, fmt.Sprintf("%*s%*s\n\n", 24, "UNITED STATES", 36, "SOVIET UNION"), time.Millisecond*5)

	bumpLine(msgr, 2)

	animateMessage(msgr, "WHICH SIDE DO YOU WANT?")
	bumpLine(msgr, 2)
	animateMessage(msgr, "  1.    UNITED STATES")
	bumpLine(msgr)
	animateMessage(msgr, "  2.    SOVIET UNION")
	bumpLine(msgr, 2)
	animateMessage(msgr, "PLEASE CHOOSE ONE: ")

	if msgr.ScanSupport() {
		_ = scan()
		handleInput(msgr, GLOBAL_THERMONOCULEAR_WAR, "")
	}

}

func strikeCommands(msgr Messenger) {
	msgr.UpdateState(STRIKE_COMMAND)
	msgr.Send(CLEAR)
	setCursorPosition(0, 0, msgr)

	awaitingMsg := "AWAITING FIRST STRIKE COMMAND"
	animateMessage(msgr, awaitingMsg)
	bumpLine(msgr)
	animateMessage(msgr, strings.Repeat(UNDERLINE, len(awaitingMsg)))

	bumpLine(msgr, 2)
	animateMessage(msgr, "PLEASE LIST PRIMARY TARGETS BY")
	bumpLine(msgr)
	animateMessage(msgr, "CITY AND/OR COUNTY NAME:")
	bumpLine(msgr, 2)

	if msgr.ScanSupport() {
		city := scan()
		handleInput(msgr, STRIKE_COMMAND, city)
	}
}

func strikeCommandDos(msgr Messenger) {
	msgr.UpdateState(STRIKE_COMMAND_2)
	bumpLine(msgr)
	if msgr.ScanSupport() {
		city := scan()

		//TODO update LOGON to next step once ready =)
		handleInput(msgr, LOGON, city)
	}
}

func notRecognized(msgr Messenger) {
	bumpLine(msgr)
	animateMessage(msgr, "IDENTIFICATION NOT RECOGNIZED BY SYSTEM\n")
	animateMessage(msgr, "--CONNECTION TERMINATED--")
	bumpLine(msgr)
	msgr.Disconnect()
	logon(msgr)
}

func scan() string {
	scanner := bufio.NewScanner(os.Stdin)
	var input string
	// Read a full line of text, including spaces
	if scanner.Scan() {
		input = scanner.Text()
	} else if err := scanner.Err(); err != nil {
		panic(err)
	}

	return input
}

func setCursorPosition(row, col int, msgr Messenger) {
	// Use ANSI escape code to set the cursor position.
	msgr.Send(fmt.Sprintf("\033[%d;%df", row+1, col+1))
}

func getTerminalWidth() int {
	// Default to 80 which seems to be standard width when a terminal opens
	width := 80

	width, _, err := term.GetSize(0)
	if err != nil {
		return width
	}
	return width
}

// Not the most elegant but a way to bump line and keep track of row without
// Perhaps I am not a fan of passing "0" in each time #GetOverIt
func bumpLine(msgr Messenger, n ...int) {
	numOfTimes := len(n)
	if numOfTimes == 0 {
		numOfTimes = 1
	} else {
		numOfTimes = n[0]
	}
	for i := 0; i < numOfTimes; i++ {
		msgr.Send("\n")
		currentRow++
	}
}
