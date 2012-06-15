var _ = require('underscore')._;

module.exports = function (config, debug, mainOverride) {
    var lines = [];
    var main = mainOverride || config.main;

    if (!config.hosts[main]) {
        throw new Error('Configuration does not contain information about host "' + main + '"');
    }

    if (debug) {
        lines.push('spawn ssh -v -F ' + config.sshConfig + ' ' + main);
    } else {
        lines.push('log_user 0');
        lines.push('spawn ssh -F ' + config.sshConfig + ' ' + main);
    }

    lines.push('expect {');
    _.each(config.hosts, function (c, name) {
        if (c.user && c.password) {
            lines.push('  # ' + name);
            lines.push('  "' + c.user + '@' + c.addr + '" {');
            lines.push('    send "' + c.password + '\\n";');
            lines.push('    exp_continue;');
            lines.push('  }');
        }
    });

    lines.push('  -re "(%|\\\\$|#) ?$" {');
    lines.push('    send_user "\\nThe login sequence seems to have worked.\\n\\n";');
    if (config.forwards) {
        lines.push('    send_user "The following forwardings have been set up for you:\\n\\n";');
        _.each(config.forwards, function (fs, descr) {
            lines.push('    send_user "' + descr + ':\\n";');
            fs.forEach(function (f) {
                lines.push('    send_user "   ' + f.from + ' -> ' + f.to + '\\n";');
            });
            lines.push('    send_user "\\n";');
        });
    }
    lines.push('    send "\\r";');
    lines.push('    interact;');
    lines.push('  }');
    lines.push('  "Permission denied" {');
    lines.push('    send_user "Permission denied, failed to set up tunneling.\n";');
    lines.push('    exit 2;');
    lines.push('  }');
    lines.push('}');

    return lines.join('\n') + '\n';
}
