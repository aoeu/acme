/*
An IRC client for acme.

Usage:

	IRC [options] <server>[:<port>]

The options are:

	-d	Enable debugging
	-f	Full name
	-n	Nickname (username)
	-p	Password
	-u	A utility program to send recieved messages via its standard input

Run "IRC" without any arguments to get a reminder of the above.

Once started, IRC will display a "server" window with a tag named "/irc/<server>"
and some of the usual acme commands, plus a "Chat" command. The body of the
server window contains messages from the IRC server, and can be used to send
raw IRC commands to the server by typing them at the ">" prompt and then typing the Enter
key.

The server window's Chat command takes a chatroom or user's name as its argument, and is
executed with mouse button 2, as usual in acme. When executed, a new window will
appear for the chatroom, named "/irc/<server>/<room>". The body of the window
will first contain the list of users in the room and the room's topic, and as IRC recieves
messages for the room, they will be added to the body, tagged with the sender's name
in angle brackets. If no one has sent any messages for five minutes, IRC will add a
timestamp to the body of the chat window. Like the server window, messages can be sent
to the room by typing them at the ">" prompt and then typing the Enter key. IRC
supports one conventional command message: /me.

Other IRC-specific tag commands:

	Who
		Prints the list of users in the room

	Nick <name>
		Changes your nickname to the given <name>
*/
package main
