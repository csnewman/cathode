import 'package:cathode_tv/app.dart';
import 'package:cathode_tv/widgets/screen.dart';
import 'dart:html' as html;

void main() {
  var screen = Screen();

  var lib = LibraryView();
  screen.setRoot(lib);
  screen.run();

  html.document.querySelector("#load-content")?.remove();
}
