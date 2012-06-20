"use strict";

var colors = require('colors');
var isatty = process.stdout.isTTY;
var debug = false;

function out(str) {
    console.log(isatty ? str : str.stripColors);
}

exports.enableDebug = function () {
    debug = true;
};

exports.debug = function (str) {
    if (debug) {
        out('debug'.magenta.bold + ' - ' + str);
    }
};

exports.fatal = function (str) {
    out('fatal'.red.bold + ' - ' + str);
    process.exit(-1);
};

exports.error = function(str) {
    out('error'.red.bold + ' - ' + str);
};

exports.warning = function(str) {
    out('warning'.yellow.bold + ' - ' + str);
};

exports.ok = function(str) {
    out('ok'.green.bold + ' - ' + str);
};

exports.info = function(str) {
    out('info'.blue.bold + ' - ' + str);
};
