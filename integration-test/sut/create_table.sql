drop database if exists testdb;

create database testdb;

use testdb;

create table usertable(YCSB_KEY varchar (255) primary key,
  FIELD0 text, FIELD1 text,
  FIELD2 text, FIELD3 text,
  FIELD4 text, FIELD5 text,
  FIELD6 text, FIELD7 text,
  FIELD8 text, FIELD9 text);
