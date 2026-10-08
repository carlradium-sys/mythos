package main

import ("bufio"; "fmt"; "net"; "fatewalker/game/character"; "fatewalker/server/session/session"; "fatewalker/world")
func main(){listener,err:=net.Listen("tcp","127.0.0.1:4000");if err!=nil{panic(err)};defer listener.Close();fmt.Println("Fatewalker: Beyond the Styx listening on 127.0.0.1:4000");w:=world.NewWorld();for{conn,err:=listener.Accept();if err!=nil{fmt.Println("Connection error:",err);continue};go func(conn net.Conn){defer conn.Close();fmt.Fprint(conn,"What is your name? ");scanner:=bufio.NewScanner(conn);if !scanner.Scan(){return};name:=scanner.Text();if name==""{name="Mortal"};s:=session.New(character.New(name),conn,w);s.Run(scanner)}(conn)}}
