package com.egoodhall.nrpc;

@FunctionalInterface
  public interface ErrorHandler {
    void handle(Exception exception) throws Exception;
  }