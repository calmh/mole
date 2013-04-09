"use strict";

var _ = require('underscore');
var colors = require('colors');
var con = require('yacon');
var fs = require('fs');
var path = require('path');
var table = require('yatf');

module.exports = list;
list.help = 'List available tunnel definitions';
list.prio = 1;
list.aliases = [ 'ls' ];

function printV3List(tunnels) {
    // Build a table with information about the tunnel definitions. Basically,
    // load each of them, create a row with information and push that row to
    // the table.

    var rows = [];
    tunnels.sort(function (a, b) {
        if (a.name < b.name)
            return -1;
        else if (a.name > b.name)
            return 1;
        return 0;
    });
    tunnels.forEach(function (tunnel) {
        if (!tunnel.error) {

            // Add a flag to indicate that a tunnel definitions requires VPN
            // (i.e. vpnc must be installed).

            var opts = '';
            if (tunnel.vpnc) {
                opts += ' (vpnc)'.magenta;
            } else if (tunnel.openconnect) {
                opts += ' (opnc)'.green;
            } else if (tunnel.socks) {
                opts += ' (socks)'.yellow;
            }

            // Generate a lists of hosts.

            var hosts = tunnel.hosts.join(', ');
            if (tunnel.localOnly) {
                hosts = '(local forward)'.grey;
            }

            rows.push([ tunnel.name.blue.bold , tunnel.description + opts, hosts ]);
        } else {
            rows.push([ tunnel.name.red.bold, tunnel.error, '-' ]);
        }
    });

    // Format the table using the specified headers and the rows from above.

    table([ 'TUNNEL', 'DESCRIPTION', 'HOSTS' ], rows, { underlineHeaders: true });
}

function printList(tunnels) {
    con.warning('Talking to a v2 server, limited list functionality.');
    var names = tunnels.map(function (t) { return t.name.replace(/\.ini$/, ''); });
    names.sort();
    names.forEach(function (n) {
        console.log(n.blue.bold);
    });
}

function list(opts, state) {
    // Get a sorted list of all tunnels.

    con.debug('Getting tunnel list');
    state.client.list(function (tunnels, proto) {
        if (proto === '3')
            printV3List(tunnels);
        else
            printList(tunnels);
    });
}
