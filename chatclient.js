var net = require('net');
const readline = require('readline');
const readlineSync = require('readline-sync');

if (process.argv.length !== 4) {
    console.log("Usage: node %s <host> <port>", process.argv[1]);
    process.exit(1);
}

var host = process.argv[2];
var port = process.argv[3];

if (host.length > 253 || port.length > 5) {
    console.log("Invalid host or port. Try again!\nUsage: node %s <host> <port>", process.argv[1]);
    process.exit(1);
}

var client = new net.Socket();
console.log("Simple chat client developed by Megan Coyne");
console.log("Connecting to: %s:%s", host, port);

client.connect(port, host, connected);

function connected() {
    console.log("Connected to: %s:%s", client.remoteAddress, client.remotePort);
    loginsync();
    // console.log("loginsync() has completed.");

    const keyboard = readline.createInterface({
        input: process.stdin,
        output: process.stdout
    });

    client.on("data", function (data) {
        console.log("Received: " + data);
        // console.log(typeof data);
        const message = data.toString().trim();
        if (message.startsWith("You have been logged out")) {
            console.log("Exiting client...");
            client.destroy();
            process.exit(0);
        } else if (message.startsWith("Authentication")) {
            loginsync()
        }
    });

    keyboard.on('line', (input) => {
        if (input.trim() === ".exit") {
            console.log("Logging out...");
            client.write(input + "\n");
        } else if (input.startsWith("[To:")) {
            console.log(`Sending private message: ${input}`);
            client.write(input + "\n");
        } else if (input.trim() !== "") {
            console.log(`Sending broadcast message: ${input}`);
            client.write(input + "\n");
        } else {
            client.write(input + "\n");
        }
    });
    // console.log("Readline listener attached after login.");
}

client.on("error", function (err) {
    console.log("Error: " + err.message);
    process.exit(2);
});

client.on("close", function () {
    console.log("Connection closed");
    process.exit(3);
});

var username;
var password;

function loginsync() {
    username = readlineSync.question('Username: ');
    if (!inputValidated(username)) {
        console.log("Username must have at least 5 characters and not more than 20. Please try again!");
        loginsync();
        return;
    }
    password = readlineSync.question('Password: ', {
        hideEchoBack: true
    });
    if (!inputValidated(password)) {
        console.log("Password must have at least 5 characters and not more than 20. Please try again!");
        loginsync();
        return;
    }
    var login = `login ${username} ${password}\n`;
    client.write(login);
    // console.log("Type your messages. '.exit' to quit."); 

}

function inputValidated(input) {
    return input.length >= 5 && input.length <= 20;
}