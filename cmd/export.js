"use strict";

var debuggable = require('debuggable');
var con = require('../lib/console');
var tun = require('../lib/tunnel');

module.exports = exportf;
exportf.help = 'Export tunnel definition to a file';
exportf.options = {
    tunnel: { position: 1, help: 'Tunnel name', required: true },
    file: { position: 2, help: 'File name to write tunnel definition to', required: true },
};
exportf.prio = 5;
debuggable(exportf);

function exportf(opts, state) {
    var config;

    // Load and verify the tunnel.

    try {
        exportf.dlog('Loading tunnel');
        config = tun.loadByName(opts.tunnel, state.path.tunnels);
    } catch (err) {
        con.fatal(err);
    }

    // Save it out to the specified file.

    exportf.dlog('Saving to INI format');
    tun.save(config, opts.file);

    con.ok(opts.file);
}
