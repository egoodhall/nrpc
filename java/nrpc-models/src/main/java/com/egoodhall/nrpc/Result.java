package com.egoodhall.nrpc;

import java.util.Objects;
import java.util.function.Consumer;
import java.util.function.Function;
import java.util.function.Supplier;

public sealed interface Result<OK, ERR> permits Result.Ok, Result.Err {
  boolean isOk();
  boolean isErr();
  OK getOkOrThrow();
  <T extends Throwable> OK getOkOrThrow(Supplier<T> errorSupplier) throws T;
  ERR getErrOrThrow();
  <T extends Throwable> ERR getErrOrThrow(Supplier<T> errorSupplier) throws T;
  <NEW_OK> Result<NEW_OK, ERR> mapOk(Function<OK, NEW_OK> mapper);
  <NEW_ERR> Result<OK, NEW_ERR> mapErr(Function<ERR, NEW_ERR> mapper);
  <NEW_OK> Result<NEW_OK, ERR> flatMapOk(Function<OK, Result<NEW_OK, ERR>> mapper);
  <NEW_ERR> Result<OK, NEW_ERR> flatMapErr(Function<ERR, Result<OK, NEW_ERR>> mapper);
  Result<OK, ERR> consume(Consumer<OK> okMapper, Consumer<ERR> errMapper);

  <O> O match(Function<OK, O> okMapper, Function<ERR, O> errMapper);

  static <OK, ERR> Result<OK, ERR> ok(OK ok) {
    return new Ok<>(ok);
  }

  static <OK, ERR> Result<OK, ERR> err(ERR err) {
    return new Err<>(err);
  }

  final class Ok<OK, ERR> implements Result<OK, ERR> {

    private final OK ok;

    private Ok(OK ok) {
      this.ok = ok;
    }

    @Override
    public boolean isOk() {
      return true;
    }

    @Override
    public boolean isErr() {
      return false;
    }

    @Override
    public OK getOkOrThrow() {
      return ok;
    }

    @Override
    public <T extends Throwable> OK getOkOrThrow(Supplier<T> errorSupplier) throws T {
      return ok;
    }

    @Override
    public ERR getErrOrThrow() {
      return getErrOrThrow(() ->
        new IncorrectResultTypeException("Error value can't be unwrapped from Ok type")
      );
    }

    @Override
    public <T extends Throwable> ERR getErrOrThrow(Supplier<T> errorSupplier) throws T {
      throw errorSupplier.get();
    }

    @Override
    public <NEW_OK> Result<NEW_OK, ERR> mapOk(Function<OK, NEW_OK> mapper) {
      return new Ok<>(mapper.apply(ok));
    }

    @Override
    public <NEW_ERR> Result<OK, NEW_ERR> mapErr(Function<ERR, NEW_ERR> mapper) {
      return new Ok<>(ok);
    }

    @Override
    public <NEW_OK> Result<NEW_OK, ERR> flatMapOk(
      Function<OK, Result<NEW_OK, ERR>> mapper
    ) {
      return mapper.apply(ok);
    }

    @Override
    public <NEW_ERR> Result<OK, NEW_ERR> flatMapErr(
      Function<ERR, Result<OK, NEW_ERR>> mapper
    ) {
      return new Ok<>(ok);
    }

    @Override
    public Result<OK, ERR> consume(Consumer<OK> okConsumer, Consumer<ERR> errConsumer) {
      okConsumer.accept(ok);
      return this;
    }

    @Override
    public <O> O match(Function<OK, O> okMapper, Function<ERR, O> errMapper) {
      return okMapper.apply(ok);
    }

    @Override
    public String toString() {
      return "Ok{" + ok + '}';
    }

    @Override
    public boolean equals(Object o) {
      if (this == o) return true;
      if (o == null || getClass() != o.getClass()) return false;
      Ok<?, ?> ok1 = (Ok<?, ?>) o;
      return Objects.equals(ok, ok1.ok);
    }

    @Override
    public int hashCode() {
      return Objects.hash(ok);
    }
  }

  final class Err<OK, ERR> implements Result<OK, ERR> {

    private final ERR err;

    private Err(ERR err) {
      this.err = err;
    }

    @Override
    public boolean isOk() {
      return false;
    }

    @Override
    public boolean isErr() {
      return true;
    }

    @Override
    public OK getOkOrThrow() {
      return getOkOrThrow(() ->
        new IncorrectResultTypeException("Ok value can't be unwrapped from Error type")
      );
    }

    @Override
    public <T extends Throwable> OK getOkOrThrow(Supplier<T> errorSupplier) throws T {
      throw errorSupplier.get();
    }

    @Override
    public ERR getErrOrThrow() {
      return err;
    }

    @Override
    public <T extends Throwable> ERR getErrOrThrow(Supplier<T> errorSupplier) throws T {
      return err;
    }

    @Override
    public <NEW_OK> Result<NEW_OK, ERR> mapOk(Function<OK, NEW_OK> mapper) {
      return new Err<>(err);
    }

    @Override
    public <NEW_ERR> Result<OK, NEW_ERR> mapErr(Function<ERR, NEW_ERR> mapper) {
      return new Err<>(mapper.apply(err));
    }

    @Override
    public <NEW_OK> Result<NEW_OK, ERR> flatMapOk(
      Function<OK, Result<NEW_OK, ERR>> mapper
    ) {
      return new Err<>(err);
    }

    @Override
    public <NEW_ERR> Result<OK, NEW_ERR> flatMapErr(
      Function<ERR, Result<OK, NEW_ERR>> mapper
    ) {
      return mapper.apply(err);
    }

    @Override
    public Result<OK, ERR> consume(Consumer<OK> okConsumer, Consumer<ERR> errOkConsumer) {
      errOkConsumer.accept(err);
      return this;
    }

    @Override
    public <O> O match(Function<OK, O> okMapper, Function<ERR, O> errMapper) {
      return errMapper.apply(err);
    }

    @Override
    public String toString() {
      return "Err{" + err + '}';
    }

    @Override
    public boolean equals(Object o) {
      if (this == o) return true;
      if (o == null || getClass() != o.getClass()) return false;
      Err<?, ?> err1 = (Err<?, ?>) o;
      return Objects.equals(err, err1.err);
    }

    @Override
    public int hashCode() {
      return Objects.hash(err);
    }
  }

  final class IncorrectResultTypeException extends RuntimeException {

    public IncorrectResultTypeException(String message) {
      super(message);
    }

    public IncorrectResultTypeException(String message, Throwable cause) {
      super(message, cause);
    }
  }
}
