package main

import (
	"fmt"
	"io"
	"net"
)

/*
ports : 1998 -> port d'écoute du reverse proxy pour la destination serveur
1999: port d'écoute pour la source client
*/
func main() {
	client, _ := net.Listen("tcp", ":1999")
	serveur, _ := net.Listen("tcp", ":1998") //serveur
	for {
		serveurConn, _ := serveur.Accept()
		fmt.Println("server connected " + serveurConn.RemoteAddr().String())

		clientConn, _ := client.Accept()
		fmt.Print("client connected ... ")
		_, werr := io.Writer(serveurConn).Write([]byte(clientConn.RemoteAddr().String() + "\x02")) //signale de gérer le client
		if werr == nil {
			//démarrage de la fonction de proxy
			fmt.Println("starting proxy for client " + clientConn.RemoteAddr().String())
			go io.Copy(serveurConn, clientConn)
			go io.Copy(clientConn, serveurConn)
		} else {
			fmt.Println(werr)
		}
		println("fin boucle")
	}

}
