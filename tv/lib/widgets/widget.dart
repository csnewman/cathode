import 'screen.dart';

abstract class Widget {
  late Screen screen;

  bool _mounted = false;

  get mounted => _mounted;

  void init(Screen screen) {
    if (_mounted) {
      return;
    }

    _mounted = true;
    this.screen = screen;

    create();
  }

  void cleanup() {
    if (!_mounted) {
      return;
    }

    _mounted = false;

    destroy();
  }

  T mount<T extends Widget>(T elem) {
    elem.init(screen);

    return elem;
  }

  void unmount(Widget elem) {
    elem.cleanup();
  }

  void create();

  void destroy();

  void update() {}
}
