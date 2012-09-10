"use strict";

var debuggable = require('debuggable');
var con = require('../lib/console');

module.exports = token;
token.help = 'Generate a new registration token';
token.prio = 5;
debuggable(token);

function token(opts, state) {
    token.dlog('Requesting new token from server');
    con.info('A token can be used only once');
    con.info('Only the most recently generated token is valid');

    // Request a token from the server. On success, the callback will be called
    // with the token and we simply print it out.

    state.client.token(function (result) {
        con.ok(result.token);
    });
}
