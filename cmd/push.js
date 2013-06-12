"use strict";

var con = require('yacon');
var fs = require('fs');
var path = require('path');

var tun = require('../lib/tunnel');

module.exports = push;
push.help = 'Send a tunnel definition to the server';
push.options = {
    'file': { position: 1, help: 'File name', required: true }
};
push.prio = 1;

function push(opts, state) {
    if (!opts.file.match(/^[0-9a-z_.\-]+\.ini$/)) {
        con.error('File name does not conform to the pattern:');
        con.error(' * must have a .ini extension,');
        con.error(' * only letters, numbers, dots, dash and underscore are allowed.');
        con.fatal('Cannot push invalid file.');
    }

    // We load the tunnel, which will cause some validation of it to happen. We
    // don't want to push files that are completely broken.

    con.debug('Testing ' + opts.file);
    try {
        tun.load(opts.file);
        con.debug('It passed validation');
    } catch (err) {
        con.fatal(err);
    }

    // We read the file to a buffer to send to the server. There should be no
    // errors here since the tunne load and check above succeeded.

    con.debug('Reading ' + opts.file);
    var data = fs.readFileSync(opts.file, 'utf-8');

    // Send the data to the server. We'll only get the callback if the upload
    // succeeds.

    var base = path.basename(opts.file);
    state.client.send(base, data, function (result) {
        con.ok('Sent ' + data.length + ' bytes');
    });
}
