"use strict";

var colors = require('colors');

var isatty = process.stdout.isTTY;
var stampEnabled = false;

function out(keyword, color, str) {
    str = keyword[color].bold + ' - '.grey + str;
    if (stampEnabled) {
        str = (new Date()).toISOString().grey + ' ' + str;
    }
    console.log(isatty ? str : str.stripColors);
}

exports.enableTimestamps = function enableTimestamps() {
    stampEnabled = true;
};

// Generate a bunch of functions for outputing tagged messages.

[ { keyword: 'error',   color: 'red' }
, { keyword: 'warning', color: 'yellow' }
, { keyword: 'ok',      color: 'green' }
, { keyword: 'info',    color: 'blue' }
, { keyword: 'debug',   color: 'magenta' } ].forEach(function (def) {
    exports[def.keyword] = function (str) { out(def.keyword, def.color, str); };
});

// 'fatal' is a special case since it causes the process to exit.

exports.fatal = function fatal(str) {
    out('fatal', 'red', str);
    process.exit(-1);
};
