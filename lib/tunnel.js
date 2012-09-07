"use strict";

var _ = require('underscore');
var fs = require('fs');
var inireader = require('inireader');
var path = require('path');

var existsSync = fs.existsSync; // Node 0.8
if (!existsSync) {
    existsSync = path.existsSync; // Node 0.6 and prior
}

var validate = require('./validate');
var con = require('./console');

module.exports = {
    name: tunnelName,
    loadByName: loadByName,
    loadFile: loadFile,
    save: saveIniTunnel
};

function tunnelName(file) {
    return path.basename(file).replace(/\.ini$/, '');
}

function loadByName(name, tunnelDefDir) {
    var local = path.join(tunnelDefDir, name) + '.ini';
    return loadFile(local);
}

function loadFile(local) {
    if (!existsSync(local)) {
        throw new Error('File "' + local + '" does not exist');
    }

    var obj = loadIniTunnel(local);
    obj.stat = fs.statSync(local);

    validate(obj);
    return obj;
}

function loadIniTunnel(name) {
    var ini = new inireader.IniReader();
    ini.load(name);
    var obj = ini.getBlock();

    var config = _.clone(obj.general);
    config.hosts = {};
    config.forwards = {};

    _.each(obj, function (val, key) {
        var m, arr;

        // Host sections look like [host host_name]
        m = key.match(/^host ([^ ]+)$/);
        if (m) {
            // SSH keys have newlines replaced by spaces
            if (val.key) {
                val.key = val.key.replace(/ /g, '\n').replace(/\nRSA\nPRIVATE\nKEY/g, ' RSA PRIVATE KEY');
            }
            config.hosts[m[1]] = val;
            return;
        }

        // Forward sections look like [forward A description here]
        m = key.match(/^(?:local)?forward +(.+)$/);
        if (m) {
            arr = [];
            _.each(val, function (to, from) {
                arr.push({ from: from, to: to });
            });
            config.forwards[m[1]] = arr;
            return;
        }

        if (key === 'vpn routes') {
            config.vpnRoutes = val;
        } else if (key !== 'general') {
            config[key] = val;
        }
    });

    return config;
}

function saveIniTunnel(config, name) {
    var ini = new inireader.IniReader();
    ini.param('general', { description: config.description, author: config.author });
    if (config.main) { ini.param('main', config.main ) };

    _.each(config.hosts, function (host, name) {
        if (host.key) {
            // The ini format doesn't handle multiline strings, so we replace newlines with spaces in ssh keys.
            host = _.clone(host);
            host.key = host.key.replace(/\n/g, ' ');
        }
        ini.param('host ' + name, host);
    });

    _.each(config.forwards, function (fwd, name) {
        var obj = {};
        fwd.forEach(function (f) {
            obj[f.from] = f.to;
        });
        ini.param('forward ' + name, obj);
    });

    if (config.vpnc) { ini.param('vpnc', config.vpnc); }
    if (config.vpnRoutes) { ini.param('vpn routes', config.vpnRoutes); }
    if (config.openconnect) { ini.param('openconnect', config.openconnect); }

    ini.write(name);
}

