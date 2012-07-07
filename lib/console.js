"use strict";

var colors = require('colors');
var iso8601 = require('iso8601');

var isatty = process.stdout.isTTY;
var debugEnabled = false;
var stampEnabled = false;

function out(keyword, color, str) {
    str = keyword[color].bold + ' - '.grey + str;
    if (stampEnabled) {
        str = iso8601.fromDate(new Date()).grey + ' ' + str;
    }
    console.log(isatty ? str : str.stripColors);
}

exports.enableDebug = function enableDebug() {
    debugEnabled = true;
};

exports.enableTimestamps = function enableTimestamps() {
    stampEnabled = true;
};

// Generate a bunch of functions for outputing tagged messages.

[ { keyword: 'error',   color: 'red' }
, { keyword: 'warning', color: 'yellow' }
, { keyword: 'ok',      color: 'green' }
, { keyword: 'info',    color: 'blue' } ].forEach(function (def) {
    exports[def.keyword] = function (str) { out(def.keyword, def.color, str); };
});

// 'debug' is a special case since it should only actually output when debug is
// enabled.

exports.debug = function debug(str) {
    if (debugEnabled) {
        out('debug', 'magenta', str);
    }
};

// 'fatal' is a special case since it causes the process to exit.

exports.fatal = function fatal(str) {
    out('fatal', 'red', str);
    process.exit(-1);
};
