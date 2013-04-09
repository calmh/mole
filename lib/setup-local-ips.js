"use strict";

var _ = require('underscore')._;
var async = require('async');
var childProcess = require('child_process');
var con = require('yacon');
var exec = childProcess.exec;
var os = require('os');
var sudo = require('sudo');

var intf;
if (process.platform === 'linux') {
    intf = 'lo';
} else if (process.platform === 'darwin' || process.platform === 'sunos' || process.platform === 'solaris') {
    intf = 'lo0';
} else {
    con.warning('Don\'t know the platform "' + process.platform + '", not setting up local IP:s for forwarding.');
}

module.exports = {
    add: add,
    remove: remove,
};

function add(config, callback) {
    if (!intf) {
        return callback(new Error('Unknown platform'));
    }

    function addIp(ip, callback) {
        con.info('Adding local IP ' + ip + ' for forwards; if asked for password, give your local (sudo) password.');
        var ifconfig = sudo(['ifconfig', intf, 'add', ip], {cachePassword: true});
        ifconfig.on('exit', function (code) {
            callback(code === 0 ? null : new Error('Failed to add IP ' + ip));
        });
    }

    var ips = [];
    _.values(config.forwards).forEach(function (fwd) {
        _.each(fwd, function (to, from) {
            var m = from.match(/^([0-9.]+):/);
            ips.push(m[1]);
        })
    });
    ips = _.uniq(ips);

    var currentIps = _.pluck(os.networkInterfaces()[intf], 'address');
    var missing = _.difference(ips, currentIps);

    // Add any missing IP:s and finally call the callback.

    async.forEachSeries(missing, addIp, callback);
}
function remove(config, callback) {
    if (!intf) {
        return callback(new Error('Unknown platform'));
    }

    function removeIp(ip, callback) {
        con.info('Removing local IP ' + ip + '; if asked for password, give your local (sudo) password.');
        var ifconfig = sudo(['ifconfig', intf, 'delete', ip], {cachePassword: true});
        ifconfig.on('exit', function (code) {
            callback(code === 0 ? null : new Error('Failed to remove IP ' + ip));
        });
    }

    var ips = [];
    _.values(config.forwards).forEach(function (fwd) {
        for (var from in fwd) {
            var m = from.match(/^([0-9.]+):/);
            ips.push(m[1]);
        }
    });

    var currentIps = _.pluck(os.networkInterfaces()[intf], 'address');
    var toRemove = _.intersect(ips, currentIps);
    toRemove = _.filter(_.uniq(toRemove), function (ip) { return !ip.match('^127\.0\.0\.(?:1)?[0-9]$'); });

    async.forEachSeries(toRemove, removeIp, callback);
}
