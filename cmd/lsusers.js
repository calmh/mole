"use strict";

var colors = require('colors');
var con = require('yacon');
var table = require('yatf');

module.exports = lsusers;
lsusers.help = 'List users (requires admin privileges)';
lsusers.options = {};
lsusers.prio = 8;

function lsusers(opts, state) {
    con.debug('Listing users');
    var rows = [];

    state.client.lsUsers(function (result) {
        for (var user in result) {
            rows.push([user.blue.bold, (new Date(result[user].created)).toISOString(), '' + !!result[user].admin, result[user].token || '']);
        }

        rows.sort();

        table(['USER', 'CREATED', 'ADMIN', 'AVAILABLE TOKEN'], rows, { underlineHeaders: true });
    });
}
