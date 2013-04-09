"use strict";

var con = require('yacon');
var tun = require('../lib/tunnel');

module.exports = exportf;
exportf.help = 'Export tunnel definition to a file';
exportf.options = {
    tunnel: { position: 1, help: 'Tunnel name', required: true },
    file: { position: 2, help: 'File name to write tunnel definition to' },
};
exportf.prio = 5;

function exportf(opts, state) {
    if (opts.file) {
        state.client.saveBin('/store/' + opts.tunnel + '.ini', opts.file, function () {
            con.ok(opts.file);
        });
    } else {
        var req = state.client.bufferedRequest({path: '/store/' + opts.tunnel + '.ini'}, function (result) {
            console.log(result.buffer);
        });
        req.end();
    }
}
