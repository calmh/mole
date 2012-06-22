"use strict";

var EventEmitter = require('events').EventEmitter;
var net = require('net');
var util = require('util');

function TrivialProxy(localHost, localPort, remoteHost, remotePort) {
    var self = this;

    // Create a server that will listen locally.

    this.server = net.createServer(function (local) {

        // We end up here when a connection is made to the local side.
        // Set up a connection to the remote side.

        var remote = net.connect(remotePort, remoteHost, function (err) {

            // An error here shouldn't bring us down completely, but it should
            // be noted and the local side should be closed.

            if (err) {
                local.end();
            } else {

                // We're connected to the remote side. Set up two pipes that simply
                // shuffle data to and fro both directions.

                local.pipe(remote);
                remote.pipe(local);
            }
        });

        // When the remote sides closes the connection, we close the local side.

        remote.on('close', function () {
            local.end();
        });

        // When the local side closes the connection, we close the remote side.

        local.on('end', function() {
            remote.end();
        });
    });

    // Set the server to listen on the local host and port as requested.

    this.server.listen(localPort, localHost, function (err) {

        // An error here is fatal; emit it for handling somewhere else.

        if (err) {
            self.emit('error', err);
        }
    });
}

util.inherits(TrivialProxy, EventEmitter);

TrivialProxy.prototype.end = function () {
    this.server.close();
};

module.exports = TrivialProxy;
