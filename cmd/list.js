"use strict";

var _ = require('underscore');
var debuggable = require('debuggable');
var fs = require('fs');
var iso8601 = require('iso8601');
var path = require('path');
var table = require('yatf');

var con = require('../lib/console');
var tun = require('../lib/tunnel');

module.exports = list;
list.help = 'List available tunnel definitions';
list.prio = 1;
list.aliases = [ 'ls' ];
debuggable(list)

function list(opts, state) {
    // Get a sorted list of all files in the tunnel directory.

    list.dlog('listing files in ' + state.path.tunnels);
    var files = fs.readdirSync(state.path.tunnels);
    files.sort();
    list.dlog('Got ' + files.length + ' files');

    // Build a table with information about the tunnel definitions. Basically,
    // load each of them, create a row with information and push that row to
    // the table.

    var rows = [];
    files.forEach(function (file) {
        var tname = tun.name(file);
        try {
            var r = tun.loadFile(path.join(state.path.tunnels, file));

            // Add a flag to indicate that a tunnel definitions requires VPN
            // (i.e. vpnc must be installed).

            var opts = '';
            if (r.vpnc) {
                opts += ' (vpnc)'.magenta;
            } else if (r.openconnect) {
                opts += ' (opnc)'.green;
            } else if (r.main && r.hosts[r.main].socks) {
                opts += ' (socks)'.yellow;
            }

            // Format the modification date.

            var mdate = iso8601.fromDate(r.stat.mtime).slice(0, 10);

            // Generate a lists of hosts. FIXME: For lots of hosts, this isn't
            // all that useful since it'll be truncated by the table formatter.

            var hosts = _.keys(r.hosts).sort().join(', ');
            if (hosts === '' && _.size(r.forwards) > 0) {
                hosts = '(local forward)'.grey;
            }

            rows.push([ tname.blue.bold , r.description + opts, mdate, hosts ]);
        } catch (err) {
            // If we couldn't load/parse the file for some reason, simply mark it as corrupt.
            rows.push([ tname.red.bold, '--Corrupt--', '--Corrupt--', '--Corrupt--' ]);
        }
    });

    // Format the table using the specified headers and the rows from above.

    table([ 'TUNNEL', 'DESCRIPTION', 'MODIFIED', 'HOSTS' ], rows, { underlineHeaders: true });
}
