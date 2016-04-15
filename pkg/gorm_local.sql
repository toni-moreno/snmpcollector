-- phpMyAdmin SQL Dump
-- version 4.3.11.1
-- http://www.phpmyadmin.net
--
-- Host: localhost
-- Generation Time: Aug 15, 2015 at 05:02 PM
-- Server version: 10.0.21-MariaDB-log
-- PHP Version: 5.5.28

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";

--
-- Database: `gorm`
--
CREATE DATABASE IF NOT EXISTS `gorm` DEFAULT CHARACTER SET utf8 COLLATE utf8_general_ci;
USE `gorm`;

-- --------------------------------------------------------

--
-- Table structure for table `contact_entries`
--

DROP TABLE IF EXISTS `contact_entries`;
CREATE TABLE IF NOT EXISTS `contact_entries` (
  `id` int(11) NOT NULL,
  `name` varchar(255) NOT NULL,
  `email` varchar(255) NOT NULL,
  `message` text,
  `mailing_address` varchar(255) NOT NULL
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `contact_entries`
--
ALTER TABLE `contact_entries`
  ADD PRIMARY KEY (`id`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `contact_entries`
--
ALTER TABLE `contact_entries`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=1;

--
-- CREATE USER and GRANT for database `gorm`
--
GRANT ALL ON `gorm`.* TO 'gorm'@'localhost' IDENTIFIED BY 'gorm';