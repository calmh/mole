var _ = require('underscore');

module.exports = function (config) {
    var lines = [];

    _.each(config.vpnc, function (val, key) {
        key = key.replace(/_/g, ' ');
        lines.push(key + ' ' + val);
    });

    return lines.join('\n') + '\n';
}

