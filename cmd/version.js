"use strict";

var libversion = require('version');

var con = require('../lib/console');

module.exports = version ;
version.help = 'Check repository for updates';
version.prio = 8;

// Fetch the latest version number for mole from the npm repository and print a
// 'time to upgrade'-message if there's a mismatch.

function version(opts, state) {
    libversion.fetch('mole', function (err, ver) {
        if (!err && ver) {
            if (ver !== state.pkg.version) {
                con.info('You are using mole v' + state.pkg.version + '; the latest version is v' + ver);
                con.info('Use "sudo npm -g update mole" to upgrade');
            } else {
                con.ok('You are using the latest version of mole');
            }
        }
    });
}
