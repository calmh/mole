"use strict";

var debuggable = require('debuggable');
var path = require('path');
var server = require('../lib/server');

module.exports = startServer;
startServer.help = 'Start a mole server instance';
startServer.options = {
    port: { abbr: 'p', help: 'Set listen port [9443]', default: '9443' },
    store: { abbr: 's', help: 'Set store directory [~/mole-store]', default: path.join(process.env.HOME, 'mole-store') },
}
startServer.prio = 9;

debuggable(startServer);
startServer.dforward(server);

function startServer(opts, state) {
    server(opts, state);
}
