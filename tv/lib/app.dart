import 'dart:html' as html;
import 'package:cathode_tv/widgets/screen.dart';
import 'widgets/widget.dart';
import 'temp.dart';

class Poster extends Widget {
  html.ImageElement img;
  html.DivElement nameBox;

  Poster()
      : img = html.ImageElement(),
        nameBox = html.DivElement() {
    img.style.borderRadius = "10px";
    img.style.position = "absolute";

    nameBox.style.position = "absolute";

    x = 0;
    y = 0;
    width = 0;
  }

  double _x = 0;
  double _y = 0;
  double _width = 0;
  bool _sel = false;
  String _path = "";
  String _name = "";

  double get x {
    return _x;
  }

  set x(double value) {
    _x = value;
    img.style.left = "${value}px";
    nameBox.style.left = "${value}px";
  }

  double get y {
    return _y;
  }

  set y(double value) {
    _y = value;
    img.style.top = "${value}px";
    _updateBox();
  }

  double get width {
    return _width;
  }

  set width(double value) {
    _width = value;
    img.style.width = "${value}px";
    nameBox.style.width = "${value}px";

    _updateBox();
  }

  void _updateBox() {
    nameBox.style.top = "${_y + (width * 1.5) + 10}px";
  }

  bool get selected {
    return _sel;
  }

  set selected(bool value) {
    _sel = value;
    img.className = _sel ? "poster-selected" : "poster";
  }

  set path(String value) {
    _path = value;
    img.src = "images/$value";
  }

  set name(String value) {
    _name = value;
    nameBox.innerText = value;
  }

  @override
  void create() {
    screen.addElement(img);
    screen.addElement(nameBox);
  }

  @override
  void destroy() {
    img.remove();
    nameBox.remove();
  }
}

class Logo extends Widget {
  html.DivElement elem;
  double _x = 0;
  double _y = 0;

  Logo() : elem = html.DivElement() {
    elem.style.position = "absolute";
    elem.className = "logo";
    elem.innerText = "Cathode";
    elem.style.fontSize = "50px";

    elem.style.width = "1000px";
    x = 0;
    y = 0;
  }

  double get x {
    return _x;
  }

  set x(double value) {
    _x = value;
    elem.style.left = "${value}px";
  }

  double get y {
    return _y;
  }

  set y(double value) {
    _y = value;
    elem.style.top = "${value}px";
  }

  @override
  void create() {
    screen.addElement(elem);
  }

  @override
  void destroy() {
    elem.remove();
  }
}

class LibraryView extends Widget {
  late Logo logo;
  late List<Poster> posters;

  @override
  void create() {
    logo = mount(Logo());
    posters = List<Poster>.empty(growable: true);
    update();
  }

  var offset = 0;

  @override
  void update() {
    // var y = 0;

    for (var p in posters) {
      p.x = 0;
      p.y = 0;
    }

    var width = screen.width - leftPad - rightBar;

    var xCount =
        ((width + posterSpacing) / (posterWidth + posterSpacing)).floor();
    var yCount = 4;
    var interPad = (width - xCount * posterWidth) / (xCount - 1);

    vertMove = xCount;

    for (var i = posters.length; i < xCount * yCount; i++) {
      posters.add(Poster());
    }

    for (var i = posters.length - 1; i >= xCount * yCount; i--) {
      unmount(posters.removeLast());
    }

    if ((position - offset) >= 3 * xCount) {
      // position -= xCount;
      offset += xCount;
    } else if (position - offset < 0) {
      offset -= xCount;
    }

    // var offset = 0;
    var id = 0;

    var y = 60.0;
    for (var j = 0; j < yCount; j++) {
      var x = leftPad;
      for (var i = 0; i < xCount; i++) {
        var poster = posters[id];

        poster.x = x;
        poster.y = y;
        poster.width = posterWidth;

        var displayId = offset + id;

        poster.selected = displayId == position;
        poster.path = imgs[displayId % imgs.length];
        poster.name = names[displayId % names.length];

        x += posterWidth + interPad;
        id++;
      }

      y += posterWidth * 1.5 + interPad + 10;
    }

    for (var p in posters) {
      mount(p);
    }
  }

  @override
  void destroy() {}
}

const leftPad = 20.0;
const rightBar = 100.0;
const posterWidth = 200.0;
const posterSpacing = 20.0;
