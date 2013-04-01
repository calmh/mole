"use strict";

var debuggable = require('debuggable');
var fs = require('fs');
var ini = require('ini')

var con = require('../lib/console');
var init = require('../lib/init');

module.exports = register;
register.help = 'Register with a mole server';
register.options = {
    server: { position: 1, help: 'Server name', required: true },
    token: { position: 2, help: 'One time registration token', required: true },
    port: { abbr: 'p', metafile: 'PORT', help: 'Set server port [9443]', default: 9443 },
};
register.prio = 5;
debuggable(register);

function register(opts, state) {
    register.dlog('Requesting registration from server ' + opts.server + ':' + opts.port);

    // Set the server and port we received from parameters in the config file,
    // and tell the server code to use them.  We don't save the config just
    // yet, though.

    state.config.server.host = opts.server;
    state.config.server.port = opts.port;
    state.client.init({ host: opts.server, port: opts.port });

    // Try to register with the server. If it fails, a fatal error will be
    // printed by the server code and the callback will never be called.  If it
    // succeeds, we'll get our certificates and the server fingerprint in the
    // callback.

    state.client.register(opts.token, function (result) {
        register.dlog('Received certificate and key from server');

        // Save the certificates and fingerprint for later.

        fs.writeFileSync(state.path.certFile, result.cert);
        fs.writeFileSync(state.path.keyFile, result.key);
        state.config.server.fingerprint = result.fingerprint;

        // Save the config file since we've verified the server and port and
        // got the fingerprint.

        fs.writeFileSync(state.path.configFile, ini.stringify(state.config));
        con.ok('Registered');
    });
}
