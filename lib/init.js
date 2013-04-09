"use strict";

var fs = require('fs');
var con = require('yacon');

module.exports = init;

// Try to read our certificates and pass them to the `server` instance. Fail
// silently if there are no certificates, since we might simply not be
// registered yet.

function readCert(state) {
    try {
        var key, cert;
        con.debug('Trying to load ' + state.path.keyFile);
        key = fs.readFileSync(state.path.keyFile, 'utf-8');
        con.debug('Trying to load ' + state.path.certFile);
        cert = fs.readFileSync(state.path.certFile, 'utf-8');
        state.client.init({ key: key, cert: cert });
    } catch (err) {
        con.debug('No certificate loaded');
    }
}

// Initialize stuff given the options from `nomnom`. This must be called early from every command callback.

function init(opts, state) {
    if (opts.debug) {
        con.enableTimestamps();
        con.enableDebug();
    }

    // The server code will check the certificate fingerprint if we have one
    // stored from before.  If not, we'll store the fingerprint on `register`,
    // thus effectively locking the client to the server it registered with and
    // preventing some tampering scenarios.

    state.client.init({
        host: state.config.server.host,
        port: state.config.server.port,
        fingerprint: state.config.server.fingerprint
    });

    readCert(state);
}
