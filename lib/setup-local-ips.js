"use strict";

var _ = require('underscore')._;
var childProcess = require('child_process');
var exec = childProcess.exec;
var os = require('os');
var spawn = childProcess.spawn;
var con = require('./console');

var intf;
if (process.platform === 'linux') {
    intf = 'lo';
} else if (process.platform === 'darwin' || process.platform === 'sunos') {
    intf = 'lo0';
} else {
    con.warning('Don\'t know the platform "' + process.platform + '", not setting up local IP:s for forwarding.');
}

module.exports = function (config, callback) {
    if (!intf) {
        return callback(false);
    }

    function addIp(ip, callback) {
        con.info('Adding local IP ' + ip + ' for forwards; if asked for password, give your local (sudo) password.');
        var ifconfig = spawn('sudo', ['ifconfig', intf, 'add', ip], { customFds: [0, 1, 2] });
        ifconfig.on('exit', callback);
    }

    function addMissingIps(exitCode) {
        if (exitCode === 0 && missing.length > 0) {
            addIp(missing.shift(), addMissingIps);
        } else {
            callback(exitCode === 0);
        }
    }

    var allForwards = _.flatten(_.values(config.forwards).concat(_.values(config.localForwards)));
    var ips = [];
    allForwards.forEach(function (f) {
        var m = f.from.match(/^([0-9.]+):/);
        ips.push(m[1]);
    });
    ips = _.uniq(ips);

    var currentIps = _.pluck(os.networkInterfaces()[intf], 'address');
    var missing = _.difference(ips, currentIps);

    // Add any missing IP:s and finally call the callback.

    addMissingIps(0);
};
