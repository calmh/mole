"use strict";

var debuggable = require('debuggable');
var fs = require('fs');
var path = require('path');

var con = require('../lib/console');
var tun = require('../lib/tunnel');
var hashSync = require('../lib/hash').sync;

module.exports = push;
push.help = 'Send a tunnel definition to the server';
push.options = {
    'file': { position: 1, help: 'File name', required: true }
};
push.prio = 1;
debuggable(push);

function push(opts, state) {
    // We load the tunnel, which will cause some validation of it to happen. We
    // don't want to push files that are completely broken.

    push.dlog('Testing ' + opts.file);
    try {
        tun.loadFile(opts.file);
        push.dlog('It passed validation');
    } catch (err) {
        con.fatal(err);
    }

    // We read the file to a buffer to send to the server. There should be no
    // errors here since the tunne load and check above succeeded.

    push.dlog('Reading ' + opts.file);
    var data = fs.readFileSync(opts.file, 'utf-8');

    // Send the data to the server. We'll only get the callback if the upload
    // succeeds.

    state.client.send(base, data, function (result) {
        con.ok('Sent ' + data.length + ' bytes');
    });
}
