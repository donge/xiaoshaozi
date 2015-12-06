//package api;

import java.io.UnsupportedEncodingException;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.Arrays;

//import https.RequestUtil;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.OutputStreamWriter;
import java.io.PrintWriter;
import java.net.URL;
import java.net.URLConnection;





public class EncryDigestUtil {

	private static String encodingCharset = "UTF-8";
	
	/**
	 * ç”Ÿæˆç­¾åæ¶ˆæ¯
	 * @param aValue  è¦ç­¾åçš„å­—ç¬¦ä¸?
	 * @param aKey  ç­¾åå¯†é’¥
	 * @return
	 */
	public static String hmacSign(String aValue, String aKey) {
		byte k_ipad[] = new byte[64];
		byte k_opad[] = new byte[64];
		byte keyb[];
		byte value[];
		try {
			keyb = aKey.getBytes(encodingCharset);
			value = aValue.getBytes(encodingCharset);
		} catch (UnsupportedEncodingException e) {
			keyb = aKey.getBytes();
			value = aValue.getBytes();
		}

		Arrays.fill(k_ipad, keyb.length, 64, (byte) 54);
		Arrays.fill(k_opad, keyb.length, 64, (byte) 92);
		for (int i = 0; i < keyb.length; i++) {
			k_ipad[i] = (byte) (keyb[i] ^ 0x36);
			k_opad[i] = (byte) (keyb[i] ^ 0x5c);
		}

		MessageDigest md = null;
		try {
			md = MessageDigest.getInstance("MD5");
		} catch (NoSuchAlgorithmException e) {

			return null;
		}
		md.update(k_ipad);
		md.update(value);
		byte dg[] = md.digest();
		md.reset();
		md.update(k_opad);
		md.update(dg, 0, 16);
		dg = md.digest();
		return toHex(dg);
	}

	public static String toHex(byte input[]) {
		if (input == null)
			return null;
		StringBuffer output = new StringBuffer(input.length * 2);
		for (int i = 0; i < input.length; i++) {
			int current = input[i] & 0xff;
			if (current < 16)
				output.append("0");
			output.append(Integer.toString(current, 16));
		}

		return output.toString();
	}

	/**
	 * 
	 * @param args
	 * @param key
	 * @return
	 */
	public static String getHmac(String[] args, String key) {
		if (args == null || args.length == 0) {
			return (null);
		}
		StringBuffer str = new StringBuffer();
		for (int i = 0; i < args.length; i++) {
			str.append(args[i]);
		}
		return (hmacSign(str.toString(), key));
	}

	/**
	 * SHAåŠ å¯†
	 * @param aValue
	 * @return
	 */
	public static String digest(String aValue) {
		aValue = aValue.trim();
		byte value[];
		try {
			value = aValue.getBytes(encodingCharset);
		} catch (UnsupportedEncodingException e) {
			value = aValue.getBytes();
		}
		MessageDigest md = null;
		try {
			md = MessageDigest.getInstance("SHA");
		} catch (NoSuchAlgorithmException e) {
			e.printStackTrace();
			return null;
		}
		return toHex(md.digest(value));

	}

	public static String testRequest(String reqUrl) throws Exception {  
	        URL url = new URL(reqUrl);  
	        URLConnection connection = url.openConnection();  
	        connection.setDoOutput(true);  
	        OutputStreamWriter out = new OutputStreamWriter(connection.getOutputStream(), "utf-8");
	        
	        out.flush();  
	        out.close();  
	        
	        String sCurrentLine;  
	        String sTotalString;  
	        sCurrentLine = "";  
	        sTotalString = "";  
	        InputStream l_urlStream;  
	        l_urlStream = connection.getInputStream();  
	        BufferedReader l_reader = new BufferedReader(new InputStreamReader(  
	                l_urlStream));  
	        while ((sCurrentLine = l_reader.readLine()) != null) {  
	            sTotalString += sCurrentLine;  
	        }  
	        return sTotalString;
	}

	public static void main(String[] args) {
		String ret;
		ret = hmacSign(args[0], digest("23a2dea7-3cd7-4fb0-ae68-55f3df58a95b"));
		System.out.println(ret);

		//System.out.println(digest("86ceb37c-9ee3-44a0-a480-29417ed4dd85"));
		/*try{
	
			String accesskey = "b193a53b-96c7-47b4-a33c-749c5a74fac3";
			String secretkey = "86ceb37c-9ee3-44a0-a480-29417ed4dd85";
			String baseURL = "https://trade.chbtc.com/api/order";
			String params = "method=order&accesskey="+accesskey+"&price=10000&amount=0.1&tradeType=0&currency=btc";
    		String secret = EncryDigestUtil.digest(secretkey);
    		String sign = EncryDigestUtil.hmacSign(params, secret);
			String url = baseURL + "?" + params + "&sign=" + sign + "&reqTime=" + System. currentTimeMillis();
			String result = testRequest(url);
			System.out.println(secret);
			System.out.println(sign);
			System.out.println(url);
			System.out.println(result);
		}catch(Exception ex){
			ex. printStackTrace();
		}*/

	}



}
