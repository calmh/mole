"use strict";

var debuggable = require('debuggable');
var https = require('https')

var con = require('../lib/console');

module.exports = version ;
version.help = 'Check repository for updates';
version.prio = 8;
debuggable(version);

// Fetch the latest version number for mole from the npm repository and print a
// 'time to upgrade'-message if there's a mismatch.

function fetchVersion(pkg, cb) {
    var options = {
        host: 'registry.npmjs.org',
        port: 443,
        path: '/' + pkg + '/latest',
        rejectUnauthorized: false
    };

    var buffer = '';
    var req = https.request(options, function (res) {
        if (res.statusCode !== 200)
            return cb(new Error('Status code ' + res.statusCode));
        res.setEncoding('utf-8');
        res.on('data', function (chunk) {
            buffer += chunk;
        });
        res.on('end', function () {
            var o = JSON.parse(buffer);
            cb(null, o.version);
        });
    });
    req.end();
}

function version(opts, state) {
    con.info('Client: mole v' + state.pkg.version);

    fetchVersion('mole', function (err, ver) {
        if (err)
            return con.error('Latest version unknown; ' + err);

        con.info('Latest: mole v' + ver);
        if (ver !== state.pkg.version) {
            con.info('Use "sudo npm -g update mole" to upgrade');
        } else {
            con.ok('You are using the latest version of mole');
        }
    });

    state.client.getPkg(function (pkg) {
        con.info('Server: mole v' + pkg.version);
    });
}
