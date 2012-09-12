"use strict";

var _ = require('underscore')._;
var fs = require('fs');
var temp = require('temp');

function forwardConfig(fwds) {
    var lines = [];
    _.each(fwds, function (fs, descr) {
        lines.push('  # ' + descr);
        fs.forEach(function (f) {
            lines.push('  LocalForward ' + f.from + ' ' + f.to);
        });
    });
    return lines;
}

module.exports = function (config) {
    var lines = [];

    _.each(config.hosts, function (c, name) {
        lines.push('Host ' + name);
        lines.push('  Compression yes');

        if (c.user) {
            lines.push('  User ' + c.user);
        }
        if (c.addr) {
            lines.push('  Hostname ' + c.addr);
        }
        if (c.port) {
            lines.push('  Port ' + c.port);
        }
        if (c.via) {
            lines.push('  ProxyCommand ssh -F ' + config.sshConfig + ' ' + c.via + ' nc -w 1800 %h %p');
        }

        if (c.key) {
            var keyFile = temp.path({suffix: '.pem'});
            fs.writeFileSync(keyFile, c.key);
            fs.chmodSync(keyFile, 384 /* 0600 octal */);
            lines.push('  IdentityFile ' + keyFile);
            lines.push('  PubkeyAuthentication yes');
        } else {
            lines.push('  PubkeyAuthentication no');
        }

        if (c.password) {
            lines.push('  PasswordAuthentication yes');
            lines.push('  KbdInteractiveAuthentication yes');
        } else {
            lines.push('  PasswordAuthentication no');
            lines.push('  KbdInteractiveAuthentication no');
        }

        if (!c.keepalive) {
            c.keepalive = 180;
        }
        var countMax = Math.max(3, Math.floor(c.keepalive / 5));
        lines.push('  ServerAliveInterval 5');
        lines.push('  ServerAliveCountMax ' + countMax);

        if (name === config.main) {
            var forwards = forwardConfig(config.forwards);
            lines.push.apply(lines, forwards);
        }
    });

    return lines.join('\n') + '\n';
};
