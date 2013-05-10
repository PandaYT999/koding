package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"koding/kontrol/helper"
)

// Gateway to messages from/to kontrold via amqp.
var listenTell *ListenTell

type ListenTell struct {
	listen chan string
	tell   chan []byte
}

const exchangeName = "infoExchange"
const channelName = "webApi"
const bindingKey = "output.webapi"

type AmqpWrapper struct {
	channel *amqp.Channel
	input   <-chan amqp.Delivery
}

func setupAmqp() (ext *AmqpWrapper) {
	connection := helper.CreateAmqpConnection()
	channel := helper.CreateChannel(connection)
	_, err := channel.QueueDeclare(channelName, false, true, false, false, nil)
	if err != nil {
		fmt.Println(err)
	}

	err = channel.QueueBind(channelName, bindingKey, exchangeName, false, nil)
	if err != nil {
		fmt.Println(err)
	}

	input, err := channel.Consume(channelName, "", true, true, false, false, nil)
	if err != nil {
		fmt.Println(err)
	}

	ext = &AmqpWrapper{channel, input}

	return
}

func (self *AmqpWrapper) Tell(cmd []byte) {
	msg := buildMessage(cmd)
	self.channel.Publish(exchangeName, "input.api", false, false, msg)
}

func (self *AmqpWrapper) Listen() <-chan amqp.Delivery {
	return self.input
}

func buildMessage(cmd []byte) (msg amqp.Publishing) {
	msg = amqp.Publishing{
		Headers:         amqp.Table{},
		ContentType:     "text/plain",
		ContentEncoding: "",
		Body:            cmd,
		DeliveryMode:    1, // 1=non-persistent, 2=persistent
		Priority:        0, // 0-9
	}

	return
}

func setupListenTell(amqp *AmqpWrapper) (listenTell *ListenTell) {
	input := make(chan string)
	output := make(chan []byte)
	listenTell = &ListenTell{input, output}

	go listenAmqp(amqp, input)
	go tellAmqp(amqp, output)

	return
}

func listenAmqp(amqp *AmqpWrapper, input chan string) {
	for msg := range amqp.Listen() {
		input <- string(msg.Body)
	}
}

func tellAmqp(amqp *AmqpWrapper, output chan []byte) {
	for cmd := range output {
		amqp.Tell(cmd)
	}
}

func (listenTell *ListenTell) Listen() string {
	return <-listenTell.listen
}

func (listenTell *ListenTell) Tell(data []byte) {
	listenTell.tell <- data
}
