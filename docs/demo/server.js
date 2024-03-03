const express = require('express');
const http = require('http');
const socketIO = require('socket.io');
const app = express();
const server = http.createServer(app);
const io = socketIO(server);
const kafka = require('kafka-node');
const Consumer = kafka.Consumer;
const client = new kafka.KafkaClient({ kafkaHost: 'localhost:9092' }); // Replace with your Kafka broker address
const consumer = new Consumer(
    client,
    [{ topic: 'events.wakemeup'}],
    { autoCommit: true }
);

consumer.on('message', function (message) {
    // Send the message to the frontend via WebSocket
    console.log("message ---->", message);
    io.emit('alert', message.value);
});

consumer.on('error', function (err) {
    console.log('Error:', err);
});

console.log("Running")