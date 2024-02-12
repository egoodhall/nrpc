package com.egoodhall.nrpc.gen.providers;

import com.egoodhall.nrpc.gen.generators.FileGenerator;
import com.google.inject.Inject;
import com.google.inject.Provider;
import com.google.protobuf.compiler.PluginProtos.CodeGeneratorResponse.File;
import com.squareup.javapoet.JavaFile;
import java.nio.file.Path;
import java.util.List;
import java.util.Set;

public class GeneratedFilesProvider implements Provider<List<File>> {

  private final Set<FileGenerator> generators;

  @Inject
  public GeneratedFilesProvider(Set<FileGenerator> generators) {
    this.generators = generators;
  }

  @Override
  public List<File> get() {
    return generators
      .stream()
      .map(FileGenerator::generate)
      .flatMap(List::stream)
      .map(this::toFile)
      .toList();
  }

  private File toFile(JavaFile javaFile) {
    return File
      .newBuilder()
      .setName(getFileName(javaFile))
      .setContent(javaFile.toString())
      .build();
  }

  private String getFileName(JavaFile javaFile) {
    return Path
      .of(javaFile.packageName.replaceAll("\\.", "/"), javaFile.typeSpec.name + ".java")
      .toString();
  }
}
