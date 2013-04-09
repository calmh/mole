"use strict";

var con = require('yacon');

module.exports = pull;
pull.help = '<deprecated>';
pull.options = { };
pull.prio = 1;

function pull(opts, state) {
    con.info('mole pull is no longer necessary');
}
