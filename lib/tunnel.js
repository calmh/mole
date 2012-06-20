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
    load: loadTunnel,
    save: saveIniTunnel
};

function tunnelName(file) {
    return path.basename(file).replace(/\.ini$/, '');
}

function loadTunnel(name, tunnelDefDir) {
    var local, stat, obj;

    if (existsSync(name)) {
        // Obviously a file name already
        local = name;
    } else if (tunnelDefDir) {
        // Unqualified names should be in the tunnel dir
        local = path.join(tunnelDefDir, name);
    }

    if (!local) {
        throw new Error('Can not find a tunnel file for "' + name + '" (1)');
    }

    if (!name.match(/(\.ini)$/)) {
        // No extension given, find the file
        if (existsSync(local + '.ini')) {
            local = local + '.ini';
        }
    }

    if (!existsSync(local)) {
        throw new Error('Can not find a tunnel file for "' + name + '" (2)');
    }

    if (local.match(/\.ini$/)) {
        obj = loadIniTunnel(local);
        obj.stat = fs.statSync(local);
    } else {
        throw new Error('Unknown format config ' + local);
    }

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
    config.vpnc = obj.vpnc;
    config.vpnRoutes = obj['vpn routes'];

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
        m = key.match(/^forward +(.+)$/);
        if (m) {
            arr = [];
            _.each(val, function (to, from) {
                arr.push({ from: from, to: to });
            });
            config.forwards[m[1]] = arr;
            return;
        }
    });

    return config;
}

function saveIniTunnel(config, name) {
    var ini = new inireader.IniReader();
    ini.param('general', { description: config.description, author: config.author, main: config.main });

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

    ini.write(name);
}

