class Province {
  final String code;
  final String name;

  Province({required this.code, required this.name});

  @override
  bool operator ==(Object other) =>
      identical(this, other) || other is Province && code == other.code;

  @override
  int get hashCode => code.hashCode;
}

class Ward {
  final String code;
  final String name;
  final String provinceCode;

  Ward({required this.code, required this.name, required this.provinceCode});

  @override
  bool operator ==(Object other) =>
      identical(this, other) || other is Ward && code == other.code;

  @override
  int get hashCode => code.hashCode;
}
