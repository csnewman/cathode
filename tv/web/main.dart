import 'package:cathode_tv/app.dart';
import 'package:cathode_tv/widgets/screen.dart';

void main() {
  var screen = Screen();

  var lib = LibraryView();
  screen.setRoot(lib);
  screen.run();
}
