var _ = require('underscore')._;
var childProcess = require('child_process');
var exec = childProcess.exec;
var spawn = childProcess.spawn;

module.exports = function (config, cli, options, callback) {
    var missing = [];
    var ipMap = {};
    var ips;

    function addIp(ip, callback) {
        cli.info('Adding local IP ' + ip + ' for forwards; if asked for password, give your local (sudo) password.');
        var ifconfig = spawn('sudo', ['ifconfig', 'lo0', 'add', ip], { customFds: [0, 1, 2] });
        ifconfig.on('exit', callback);
    }

    function addMissingIps(exitCode) {
        if (exitCode === 0 && missing.length > 0) {
            addIp(missing.shift(), addMissingIps);
        } else {
            callback(exitCode === 0);
        }
    }

    _.each(config.forwards, function (fs, name) {
        fs.forEach(function (f) {
            var m = f.from.match(/^([0-9.]+):/);
            if (m) {
                ipMap[m[1]] = true;
            }
        });
    });
    ips = _.keys(ipMap);

    exec('ifconfig lo0', function (error, stdout, stderr) {
        ips.forEach(function (ip) {
            if (!stdout.match(new RegExp('\\s' + ip.replace('.', '\\.') + '\\s'))) {
                missing.push(ip);
            }
        });

        addMissingIps(0);
    });
}
