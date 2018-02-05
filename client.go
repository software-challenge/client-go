package main

import (
	"encoding/xml"
	"fmt"
	"github.com/droundy/goopt"
	"io"
	"net"
)

type Room struct {
	Id string `xml:"roomId"`
}

type Joined struct {
	Id string `xml:"roomId"`
}

type WelcomeMessage struct {
	Color string `xml:"color,attr"`
}

type Memento struct {
	State State `xml:"state"`
}

type Card struct {
	CardType string `xml:"type"`
}

type Player struct {
	DisplayName string `xml:"displayName,attr"`
	Color       string `xml:"color,attr"`
	Index       int    `xml:"index,attr"`
	Carrots     int    `xml:"carrots,attr"`
	Salads      int    `xml:"salads,attr"`
	Cards       []Card `xml:"cards"`
}

type State struct {
	RedPlayer  Player `xml:"red"`
	BluePlayer Player `xml:"blue"`
	Board      Board  `xml:"board"`
}

type Field struct {
	FieldType string `xml:"type,attr"`
	Index     int    `xml:"index,attr"`
}

type Board struct {
	Fields []Field `xml:"fields"`
}

func Process(r io.Reader, w io.Writer) error {
	d := xml.NewDecoder(r)
	var myColor string
	var currentState State
	var roomId string
	for {
		v, err := d.Token()
		if err != nil {
			return err
		}

		switch t := v.(type) {

		case xml.StartElement:
			switch t.Name.Local {
			case "data":
				var class string = ""
				for _, v := range t.Attr {
					if v.Name.Local == "class" {
						class = v.Value
						break
					}
				}
				switch class {
				case "memento":
					data := new(Memento)
					err := d.DecodeElement(data, &t)
					if err != nil {
						return err
					}
					currentState = data.State
					fmt.Printf("got memento %#v\n", data)
				case "welcomeMessage":
					data := new(WelcomeMessage)
					err := d.DecodeElement(data, &t)
					if err != nil {
						return err
					}
					myColor = data.Color
					fmt.Printf("we have color %s\n", myColor)
				case "sc.framework.plugins.protocol.MoveRequest":
					// move to next carrot field
					var us Player
					var opponent Player
					if myColor == "red" {
						us = currentState.RedPlayer
						opponent = currentState.BluePlayer
					} else {
						us = currentState.BluePlayer
						opponent = currentState.RedPlayer
					}
					var nextCarrotFieldIndex int
					for _, v := range currentState.Board.Fields[us.Index+1 : 64] {
						fmt.Printf("%d : %s\n", v.Index, v.FieldType)
						if v.FieldType == "CARROT" && opponent.Index != v.Index {
							nextCarrotFieldIndex = v.Index
							break
						}
					}
					fmt.Printf("we are at %d\n", us.Index)
					fmt.Printf("next carrot field is at index %d\n", nextCarrotFieldIndex)
					distance := nextCarrotFieldIndex - us.Index

					if (distance*(distance+1))/2 > us.Carrots {
						// not enough carrots to move forward
						fmt.Println("will fall back to hedgehog field")
						io.WriteString(w, fmt.Sprintf("<room roomId=\"%s\"><data class=\"move\"><fallBack order=\"0\" /></data></room>", roomId))
					} else {
						fmt.Printf("will advance %d fields\n", distance)
						io.WriteString(w, fmt.Sprintf("<room roomId=\"%s\"><data class=\"move\"><advance order=\"0\" distance=\"%d\" /></data></room>", roomId, distance))

					}

				default:
					fmt.Printf("got data of class %s\n", class)
				}
			case "joined":
				for _, v := range t.Attr {
					if v.Name.Local == "roomId" {
						roomId = v.Value
						break
					}
				}
				fmt.Printf("joined room %s\n", roomId)

			default:
				//fmt.Printf("got xml start tag %s\n", t.Name.Local)
			}

		case xml.EndElement:
			//fmt.Printf("got xml end tag %s\n", t.Name.Local)
		}
	}
}

func main() {
	goopt.Description = func() string {
		return "Software Challenge client"
	}
	goopt.Version = "0.1"
	goopt.Summary = "Software Challenge client"
	host := goopt.String([]string{"-h", "--host"}, "localhost", "hostname or IP address of the game server")
	port := goopt.Int([]string{"-p", "--port"}, 13050, "port of the game server")
	reservation := goopt.String([]string{"-r", "--reservation"}, "", "reservation id for the game to join (if empty, new game will be joined)")
	goopt.Parse(nil)

	con, err := net.Dial("tcp", fmt.Sprintf("%s:%d", *host, *port))
	if err != nil {
		panic("could not connect to server")
	}
	fmt.Println("connected to server")
	//	d := xml.NewDecoder(con)
	io.WriteString(con, "<protocol>")
	if *reservation == "" {
		io.WriteString(con, "<join gameType=\"swc_2018_hase_und_igel\"/>")
	} else {
		io.WriteString(con, fmt.Sprintf("<joinPrepared reservationCode=\"%s\"/>", *reservation))
	}
	err = Process(con, con)
	fmt.Printf("Err: %#v\n", err)
}
