"use strict";

var _ = require('underscore')._;

module.exports = function expectConfig(config, debug) {
    var lines = [];

    if (!config.hosts[config.general.main]) {
        throw new Error('Configuration does not contain information about host "' + config.general.main + '"');
    }

    if (!debug) {
        lines.push('log_user 0');
    }

    lines.push('set timeout 30');

    if (debug) {
        lines.push('spawn ssh -v -F ' + config.sshConfig + ' ' + config.general.main);
    } else {
        lines.push('spawn ssh -F ' + config.sshConfig + ' ' + config.general.main);
    }

    lines.push('expect {');
    _.each(config.hosts, function (c, name) {
        if (c.user && c.password) {
            lines.push('  # ' + name);
            lines.push('  "' + c.user + '@' + c.addr + '" {');
            lines.push('    send "' + c.password + '\\n";');
            lines.push('    exp_continue;');
            lines.push('  }');

            if (name === config.general.main) {
                lines.push('  # ' + name + ' as only host');
                lines.push('  "Password:" {');
                lines.push('    send "' + c.password + '\\n";');
                lines.push('    exp_continue;');
                lines.push('  }');
            }
        }
    });

    if (config.hosts[config.general.main].prompt) {
        lines.push('  -re "' + config.hosts[config.general.main].prompt + '" {');
    } else {
        lines.push('  -re "(%|\\\\$|#|>)\\\\s*$" {');
    }
    lines.push('    send_user "\\nThe login sequence seems to have worked.\\n\\n";');
    if (config.forwards) {
        lines.push('    send_user "The following forwardings have been set up for you:\\n\\n";');
        _.each(config.forwards, function (fs, descr) {
            lines.push('    send_user "' + descr + ':\\n";');
            _.each(fs, function (to, from) {
                lines.push('    send_user "   ' + from + ' -> ' + to + '\\n";');
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
    lines.push('  timeout {');
    lines.push('    send_user "Unknown error, failed to set up tunneling.\n";');
    lines.push('    exit 2;');
    lines.push('  }');
    lines.push('}');

    // Propagate exit code
    lines.push('catch wait reason');
    lines.push('exit [lindex $reason 3]');

    return lines.join('\n') + '\n';
};
