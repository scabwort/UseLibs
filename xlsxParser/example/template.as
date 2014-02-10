package com.youzu.tot.config
{
	import flash.utils.ByteArray;
	import flash.utils.Dictionary;
	/**
	 * $Info
	 * 自动生成，请务修改
	 */
	public class $ClsName
	{
		static public const urlKey:String = "$TblUrl";
		$Prop
		static public function parse(bytes:ByteArray):$DicType
		{
			var dic:$DicType = new $DicType();
			while (bytes.bytesAvailable)
			{
				var vo:$ClsName = new $ClsName();
				$Reader
				$DicAdd
			}
			return dic;
		}
	}
}||/** $Info */
		public var $Prop:$Type;
		||vo.$Prop = $Reader;
				||dic[vo.$DicKey] = vo;||dic.push(vo);