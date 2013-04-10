"use strict";

var con = require('yacon');

module.exports = del;
del.help = 'Delete a tunnel definition from the server';
del.options = {
    'tunnel': { position: 1, help: 'Tunnel name', required: true }
};
del.prio = 5;
del.aliases = ['rm'];

function del(opts, state) {
    var path = '/store/' + opts.tunnel + '.ini';
    var req = state.client.bufferedRequest({path: path, method: 'DELETE'}, function (result) {
        con.ok('Deleted ' + opts.tunnel);
    });
    req.end();
}
