"use strict";

var winston = require('winston');
var path = require('path');

var MAXMB = 10;
var MAXFILES = 5;

module.exports = audit;

function audit(opts) {
    var alog = new (winston.Logger)({
        transports: [ new (winston.transports.File)({
            filename: opts.auditFile,
            level: 0,
            timestamp: true,
            maxsize: MAXMB * 1024 * 1024,
            maxFiles: MAXFILES
        })]
    });

    // debug: 0, info: 1, notice: 2, warning: 3, error: 4, crit: 5, alert: 6, emerg: 7
    alog.setLevels(winston.config.syslog.levels);
    alog.info('audit started');

    return alog;
}

