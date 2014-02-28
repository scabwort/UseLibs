package ;
import format.jpg.Data;
import format.jpg.Writer;
import format.png.Reader;
import format.png.Tools;
import haxe.io.BytesInput;
import sys.io.File;

/**
 * ...
 * @author Scott Lee
 */
class TestPng
{

	public function new() 
	{
		var file = new BytesInput(File.getBytes("pic/1/0/0000000.png"));
		var data = new format.png.Reader(file).read();
		//data
		
		var head = Tools.getHeader(data);
		
		var jpgData:Data = {
			width : head.width,
			height : head.height,
			quality : 90,
			pixels : Tools.extract32(data)
		}
		
		var out = File.write("pp.jpg");
		
		
		new format.jpg.Writer(out).write(jpgData);
		out.close();
	}
	
	static function main()
	{
		new TestPng();
	}
}