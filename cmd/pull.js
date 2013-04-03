"use strict";

var con = require('../lib/console');

module.exports = pull;
pull.help = 'Create a new user (requires admin privileges)';
pull.options = { };
pull.prio = 1;

function pull(opts, state) {
    con.info('mole pull is no longer necessary');
}
