package com.egoodhall.nrpc;

@FunctionalInterface
public interface ErrorHandler {
  void handle(Throwable exception);
}
