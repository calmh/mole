"use strict";

var path = require('path');
var server = require('../lib/server');

module.exports = startServer;
startServer.help = 'Start a mole server instance';
startServer.options = {
    port: { abbr: 'p', help: 'Set listen port [9443]', default: '9443' },
    store: { abbr: 's', help: 'Set store directory [~/mole-store]', default: path.join(process.env.HOME, 'mole-store') },
    ldapsrv: { help: 'Set LDAP server [ldap://127.0.0.1]', default: 'ldap://127.0.0.1' },
    ldapdn: { help: 'Set LDAP bind DN pattern [uid=%s,cn=users,cn=accounts,dc=example,dc=com]', default: 'uid=%s,cn=users,cn=accounts,dc=example,dc=com' },
};
startServer.prio = 9;

function startServer(opts, state) {
    server(opts, state);
}
