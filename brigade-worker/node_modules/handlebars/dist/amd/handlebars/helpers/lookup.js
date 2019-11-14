define(['exports', 'module'], function (exports, module) {
  'use strict';

  module.exports = function (instance) {
    instance.registerHelper('lookup', function (obj, field) {
      if (!obj) {
        return obj;
      }
      if (String(field) === 'constructor' && !obj.propertyIsEnumerable(field)) {
        return undefined;
      }
      return obj[field];
    });
  };
});
//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIi4uLy4uLy4uLy4uL2xpYi9oYW5kbGViYXJzL2hlbHBlcnMvbG9va3VwLmpzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiI7OzttQkFBZSxVQUFTLFFBQVEsRUFBRTtBQUNoQyxZQUFRLENBQUMsY0FBYyxDQUFDLFFBQVEsRUFBRSxVQUFTLEdBQUcsRUFBRSxLQUFLLEVBQUU7QUFDckQsVUFBSSxDQUFDLEdBQUcsRUFBRTtBQUNSLGVBQU8sR0FBRyxDQUFDO09BQ1o7QUFDRCxVQUFJLE1BQU0sQ0FBQyxLQUFLLENBQUMsS0FBSyxhQUFhLElBQUksQ0FBQyxHQUFHLENBQUMsb0JBQW9CLENBQUMsS0FBSyxDQUFDLEVBQUU7QUFDdkUsZUFBTyxTQUFTLENBQUM7T0FDbEI7QUFDRCxhQUFPLEdBQUcsQ0FBQyxLQUFLLENBQUMsQ0FBQztLQUNuQixDQUFDLENBQUM7R0FDSiIsImZpbGUiOiJsb29rdXAuanMiLCJzb3VyY2VzQ29udGVudCI6WyJleHBvcnQgZGVmYXVsdCBmdW5jdGlvbihpbnN0YW5jZSkge1xuICBpbnN0YW5jZS5yZWdpc3RlckhlbHBlcignbG9va3VwJywgZnVuY3Rpb24ob2JqLCBmaWVsZCkge1xuICAgIGlmICghb2JqKSB7XG4gICAgICByZXR1cm4gb2JqO1xuICAgIH1cbiAgICBpZiAoU3RyaW5nKGZpZWxkKSA9PT0gJ2NvbnN0cnVjdG9yJyAmJiAhb2JqLnByb3BlcnR5SXNFbnVtZXJhYmxlKGZpZWxkKSkge1xuICAgICAgcmV0dXJuIHVuZGVmaW5lZDtcbiAgICB9XG4gICAgcmV0dXJuIG9ialtmaWVsZF07XG4gIH0pO1xufVxuIl19
