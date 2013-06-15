"use strict";

var colors = require('colors');
var con = require('yacon');
var table = require('yatf');
var sprintf = require('sprint');

module.exports = lsusers;
lsusers.help = 'List users (requires admin privileges)';
lsusers.options = {};
lsusers.prio = 8;

function dateStr(d) {
    try {
    if (typeof d === 'number')
        d = new Date(d);
    return sprintf('%4d-%02d-%02d', d.getUTCFullYear(), d.getUTCMonth() + 1, d.getUTCDate());
    } catch (e) {
        return '-';
    }
}

function lsusers(opts, state) {
    con.debug('Listing users');
    var rows = [];

    state.client.lsUsers(function (result) {
        for (var user in result) {
            var created = dateStr(result[user].created);
            var seen = dateStr(result[user].seen);
            var isAdmin = '' + !!result[user].admin;
            var token = result[user].token || '';

            rows.push([user.blue.bold, created, seen, isAdmin, token]);
        }

        rows.sort();

        table(['USER', 'CREATED', 'SEEN', 'ADMIN', 'AVAILABLE TOKEN'], rows, { underlineHeaders: true });
    });
}
