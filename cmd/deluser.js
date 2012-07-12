"use strict";

var con = require('../lib/console');

module.exports = deluser;
deluser.options = { name: { position: 1, help: 'User name', required: true } };
deluser.help = 'Delete a user (requires admin privileges)';
deluser.prio = 8;

function deluser(opts, state) {
    // Delete a user from the server. As always, we only get the callback if
    // everything went well.

    con.debug('Deleting user ' + opts.name);
    state.client.delUser(opts.name, function (result) {
        con.ok('deleted');
    });
}
