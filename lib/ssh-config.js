var _ = require('underscore')._;

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
        c.user && lines.push('  User ' + c.user);
        c.addr && lines.push('  Hostname ' + c.addr);
        c.port && lines.push('  Port ' + c.port);
        c.via && lines.push('  ProxyCommand ssh ' + c.via + ' nc -w 1800 %h %p');
        if (name === config.main) {
            var forwards = forwardConfig(config.forwards);
            lines.push.apply(lines, forwards);
        }
    });
    return lines.join('\n');
}
