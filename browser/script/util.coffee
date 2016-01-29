

hex2str = (str) ->
  str = str.toString()
  hexes = str.match(/.{1,4}/g) or []
  back = ''
  j = 0
  while j < hexes.length
    back += String.fromCharCode(parseInt(hexes[j], 16))
    j++
  back


module.exports = {
  hex2str: hex2str

}
