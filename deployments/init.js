// Script de inicialización de MongoDB
db = db.getSiblingDB('message-db');
db.createCollection('messages'); 