"use strict";

var con = require('../lib/console');

module.exports = newuser;
newuser.help = 'Create a new user (requires admin privileges)';
newuser.options = {
    'name': { position: 1, help: 'User name', required: true },
    'admin': { flag: true, abbr: 'a', help: 'Create an admin user' },
};
newuser.prio = 8;

function newuser(opts, state) {
    // Create a new user on the server. If the call succeeds, we'll get the
    // callback with the one-time token for the new user.

    con.debug('Requesting user ' + opts.name);
    state.client.newUser(opts.name, function (result) {
        con.ok(result.token);
    });
}
