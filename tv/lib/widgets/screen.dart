import 'widget.dart';
import 'dart:html' as html;

var position = 0;
var vertMove = 0;

class Screen {
  late Widget root;

  double get width => html.window.innerWidth! as double;

  double get height => html.window.innerHeight! as double;

  void addElement(html.Element elem) {
    html.document.body!.children.add(elem);
  }

  void setRoot(Widget widget) {
    root = widget;
    // widget.init(this);
  }

  void run() {
    html.window.addEventListener("resize", (event) => render());
    html.window.addEventListener("keydown", (event) {
      var evt = event as html.KeyboardEvent;
      print(evt.key);

      if (evt.key == "ArrowRight") {
        position++;
        render();
      } else if (evt.key == "ArrowLeft") {
        position--;
        render();
      } else if (evt.key == "ArrowUp") {
        position -= vertMove;
        render();
      } else if (evt.key == "ArrowDown") {
        position += vertMove;
        render();
      }
    });

    root.init(this);

    render();
  }

  void render() {
    root.update();
    // root.
  }
}
